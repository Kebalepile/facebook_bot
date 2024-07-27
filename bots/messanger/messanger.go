package messanger

import (
	"bufio"
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type Bot struct {
	Name    string
	URL     string
	Quit    context.CancelFunc
	Message string
}

func (b *Bot) Run(wg *sync.WaitGroup) {
	defer wg.Done()
	log.Printf("init: %v", b.Name)

	// Load the .env variables
	envVars := b.env()

	// Get email and password from .env
	email := envVars["EMAIL"]
	password := envVars["PASSWORD"]

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false), // set headless to true for production
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.163 Safari/537.36"),
		chromedp.WindowSize(1000, 755), // Laptop size
	)
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	// defer cancel() // Removed to keep the browser open

	ctx, cancel = chromedp.NewContext(ctx)
	b.Quit = cancel

	log.Println(b.Name, "loading URL: ", b.URL)

	err := chromedp.Run(ctx,
		chromedp.Navigate(b.URL),
		b.waitVisibleAndSendKeys(`//*[@placeholder='Email address or phone number']`, email),
		b.waitVisibleAndSendKeys(`//*[@placeholder='Password']`, password),
		b.waitVisibleAndClick(`//button[text()='Log in']`),
	)
	b.error(err)

	// Keep the browser open and wait for user input
	// b.waitForUserInput()
	b.pause(10)

	// Wait for end-user to type 'continue' & press Enter before proceeding
	b.waitForContinue()

	// Handle alerts
	b.navigate_to_group(ctx)
}

// Wait for user to type 'continue' and press Enter
func (b *Bot) waitForContinue() {
	log.Printf("%v paused waiting for you command \n\nType 'continue' and press Enter to proceed...", b.Name)

	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading input:", err)
			continue
		}

		// Trim any whitespace or newline characters from the input
		input = strings.TrimSpace(input)

		// Check if the input is "continue"
		if input == "continue" {
			break
		} else {
			log.Println("Invalid input. Type 'continue' and press Enter to proceed...")
		}
	}
}

// Navigates to the group and clicks the more SVG once visible
func (b *Bot) navigate_to_group(ctx context.Context) {
	moreSvg := "M3.25 2.75a1.25 1.25 0 1 0 0 2.5h17.5a1.25 1.25 0 1 0 0-2.5H3.25zM2 12c0-.69.56-1.25 1.25-1.25h17.5a1.25 1.25 0 1 1 0 2.5H3.25C2.56 13.25 2 12.69 2 12zm0 8c0-.69.56-1.25 1.25-1.25h17.5a1.25 1.25 0 1 1 0 2.5H3.25C2.56 21.25 2 20.69 2 20z"

	b.pause(10)
	log.Println("navigating to group")
	log.Println("entering group name to the search group, search form")

	selector := `input[placeholder="Search groups"]`
	searchGroup := b.env()["SEARCH_GROUP"]
	err := chromedp.Run(ctx,
		b.clickSvgParentElementByPath(moreSvg),
		chromedp.Evaluate(`document.querySelectorAll('span').forEach(element => {
			if (element.textContent.trim() === 'Groups') {
				element.click();
			}
		});`, nil),
		b.waitVisibleAndSendKeys(selector, searchGroup), // Send the search query
		chromedp.Sleep(3*time.Second),
		chromedp.SendKeys(selector, "\n"), // Simulate pressing the Enter key

		// Pause for 10 seconds before executing the next action
		chromedp.Sleep(10*time.Second),

		// Ensure the group is public, find and click the first matching element
		chromedp.ActionFunc(func(ctx context.Context) error {
			var result string
			jsCode := fmt.Sprintf(`
				(function() {
					// Check for 'Public' in any span element
					let publicGroup = Array.from(document.querySelectorAll('span')).find(span => 
						span.textContent.toLowerCase().includes('public') && span.textContent.toLowerCase().includes('members')
					);
		
					if (!publicGroup) {
						// console.log('Group is not public');
						return 'Group is not public';
					}
		
					// Find and click the <a> tag that matches the group name
					let targetText = "%s".toLowerCase(); // This will be dynamic in your actual use case
					let groupElement = publicGroup.closest('[role="feed"]').querySelectorAll('a[aria-hidden="true"]');
					let clicked = false;
		
					for (let i = 0; i < groupElement.length; i++) {
						let element = groupElement[i];
						if (element.textContent.toLowerCase().includes(targetText)) {
							element.click(); // Click the first matching element
							clicked = true;
							// console.log('Group clicked');
							return 'Group clicked';
						}
					}
		
					if (!clicked) {
						console.log('No matching group found');
						return 'No matching group found';
					}
				})();
			`, searchGroup)
			err := chromedp.Evaluate(jsCode, &result).Do(ctx)
			if err != nil {
				log.Printf("Error evaluating JS: %v", err)
				return err
			}
			log.Println(result)
			return nil
		}),
	)
	b.error(err)
	b.waitForUserInput()
}

// Click an SVG element by matching its path attribute after waiting for it to be visible
func (b *Bot) clickSvgParentElementByPath(pathValue string) chromedp.Tasks {

	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			var nodes []*cdp.Node
			err := chromedp.Nodes(`svg path`, &nodes, chromedp.ByQueryAll).Do(ctx)
			if err != nil {
				return err
			}

			if len(nodes) > 0 {
				for _, node := range nodes {
					pathAttr := node.AttributeValue("d")
					if pathAttr == pathValue {
						parent := node.Parent.Parent // Get the grandparent element
						if parent != nil {
							// Click the grandparent element
							return chromedp.Click(parent.FullXPath()).Do(ctx)
						}
					}
				}
			}
			return nil
		}),
	}
}

func (b *Bot) waitVisibleAndSendKeys(selector, keys string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.WaitVisible(selector, chromedp.BySearch),
		chromedp.SendKeys(selector, keys, chromedp.BySearch),
	}
}

func (b *Bot) waitVisibleAndClick(selector string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.WaitVisible(selector, chromedp.BySearch),
		chromedp.Click(selector, chromedp.BySearch),
	}
}

func (b *Bot) waitForUserInput() {
	log.Println("Type 'e' or 'exit' and press Enter to exit...")

	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading input: %v", err)
			continue
		}

		// Trim any whitespace or newline characters from the input
		input = strings.TrimSpace(input)

		// Check if the input is "e" or "exit"
		if input == "e" || input == "exit" {
			break
		} else {
			log.Println("Invalid input. Type 'e' or 'exit' and press Enter to exit...")
		}

	}

	b.quit()
}

// Read .env variables to be used
func (b *Bot) env() map[string]string {
	variables, err := godotenv.Read()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	return variables
}

// pauses spider for given duration
func (b *Bot) pause(second int) {
	time.Sleep(time.Duration(second) * time.Second)
}

// closes chromedp browser instance
func (b *Bot) quit() {
	log.Printf("%v done.", b.Name)
	b.Quit()
}

func (b *Bot) error(err error) {
	if err != nil {
		log.Println("*************************************")
		log.Printf(" %v Error:", b.Name)
		log.Println(err.Error())
		log.Printf(" %v please restart bot", b.Name)
		log.Println("*************************************")
		log.Fatal(err)
	}
}
