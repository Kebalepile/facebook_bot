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
	log.Printf("%v paused waiting for you command. \nType 'continue' and press Enter to proceed...", b.Name)

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
    moreSvg := "M3.25 2.75a1.25 1.25 0 1 0 0 2.5h17.5a1.25 1.25 0 1 0 0-2.5H3.25zM2 12c0-.69.56-1.25 1.25-1.25h17.5a1.25 1.25 0 1 1 0 2.5H3.25C2.56 13.25 2 12zm0 8c0-.69.56-1.25 1.25-1.25h17.5a1.25 1.25 0 1 1 0 2.5H3.25C2.56 21.25 2 20.69 2 20z"

    b.pause(10)
    log.Println("Navigating to group")
    log.Println("Entering group name to the search group, search form")

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
                function findGroupSection() {
                    return Array.from(document.querySelectorAll('span')).find(span =>
                    span.textContent.toLowerCase().includes('public') && (span.textContent.toLowerCase().includes('members') || span.textContent.toLowerCase().includes('people'))
                    );
                }

                let publicGroup = findGroupSection();

                if (!publicGroup) {
                    return 'Group is not public';
                }

                let targetText = "%s".toLowerCase(); // This will be dynamic in your actual use case
                let groupElement = publicGroup.closest('[role="feed"]').querySelectorAll('a[aria-hidden="true"]');
                let clicked = false;

                for (let i = 0; i < groupElement.length; i++) {
                    let element = groupElement[i];
                    if (element.textContent.toLowerCase().includes(targetText)) {
                        element.click(); // Click the first matching element
                        clicked = true;
                        return 'Group clicked';
                    }
                }

                if (!clicked) {
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
        chromedp.Sleep(10*time.Second),
       b.addFriendsFromNewToGroupSection(ctx),
    )
    b.error(err)

    b.waitForUserInput()
}

func (b *Bot) addFriendsFromNewToGroupSection(ctx context.Context) chromedp.ActionFunc  {

	// Ask the user if they want to send friend requests
	fmt.Print("Do you want to send friend requests? (yes/y or no/n): ")
	
	var response string
	fmt.Scanln(&response)
	
	response = strings.ToLower(strings.TrimSpace(response))
	
	if response == "yes" || response == "y" {
		return chromedp.ActionFunc(func(ctx context.Context) error {
			// Click the "people" element
			err := b.clickElementWithText(ctx, "people")
			if err != nil {
				log.Printf("Error clicking element: %v", err)
				return fmt.Errorf("error clicking 'people' element: %w", err)
			}
	
			// Execute the JavaScript to add friends from the "New to the group" section
			var result map[string]interface{}
			err = chromedp.Evaluate(`
				function addFriendsFromNewToGroupSection() {
					let adminNames = [];
					let clicked = 0;

					function findAndClickAddFriendButtons() {
						let newGroupSection = Array.from(document.querySelectorAll('div'))
							.find(div => div.textContent.includes('New to the group'));

						if (newGroupSection) {
							let addFriendButtons = newGroupSection.querySelectorAll('[aria-label="Add friend"]');
							for (let i = 0; i < addFriendButtons.length; i++) {
								let card = addFriendButtons[i].closest('div[role="listitem"]');
								if (card) {
									let nameElement = card.querySelector('a[role="link"]');
									let name = nameElement ? nameElement.textContent.trim() : null;

									if (card.textContent.includes('Admin')) {
										if (name) {
											adminNames.push(name);
										}
									} else {
										if (name && adminNames.includes(name)) {
											// Skipping button because the name is flagged as an Admin
										} else {
											// Clicking button within 'New to the group' section
											addFriendButtons[i].click();
											clicked++;
										}
									}
								}
							}
							return true;
						} else {
							return false;
						}
					}

					function scrollToBottom() {
						window.scrollTo(0, document.body.scrollHeight);
					}

					function loadAndClick() {
						if (findAndClickAddFriendButtons()) {
							setTimeout(() => {
								scrollToBottom();
								setTimeout(() => {
									if (findAndClickAddFriendButtons()) {
										loadAndClick();
									} else {
										console.log({
											status: 'Success',
											clickedCount: clicked,
											adminNames: adminNames
										});
									}
								}, 2000); // Wait for 2 seconds after scrolling before clicking
							}, 2000); // Wait for 2 seconds before scrolling again
						} else {
							console.log({
								status: 'New to the group section not found',
								clickedCount: clicked,
								adminNames: adminNames
							});
						}
					}

					loadAndClick();
				}

				addFriendsFromNewToGroupSection();


			`, &result).Do(ctx)
	
			if err != nil {
				log.Printf("Error executing JavaScript: %v", err)
				return fmt.Errorf("error executing JavaScript: %w", err)
			}
	
			// Log the result of the JavaScript execution
			log.Printf("Add Friends Result: %v", result)
	
			// Return an error if the result indicates failure
			// if result["status"] != "Success" {
			// 	return fmt.Errorf("failed to add friends: %v", result["status"])
			// }
	
			return nil
		})
	}
	return nil
}


// clickElementWithText clicks on the element containing the text 'People', case-insensitively
func (b *Bot) clickElementWithText(ctx context.Context, text string) error {
	var result string
	jsCode := fmt.Sprintf(`
		(function() {
			// Convert target text to lower case
			let targetText = "%s".toLowerCase();
			
			// Find all elements (e.g., span, div, etc.)
			let elements = document.body.querySelectorAll('*');
			let clicked = false;

			for (let element of elements) {
				if (element.textContent.toLowerCase().includes(targetText)) {
					// Find the closest <a> tag ancestor
					let anchorElement = element.closest('a');
					if (anchorElement) {
						anchorElement.click(); // Click the <a> element
						clicked = true;
						console.log('Anchor element clicked');
						return 'Anchor element clicked';
					}
				}
			}

			if (!clicked) {
				console.log('No matching element found');
				return 'No matching element found';
			}
		})();
	`, text)

	// Evaluate the JavaScript code
	err := chromedp.Evaluate(jsCode, &result).Do(ctx)
	if err != nil {
		log.Printf("Error evaluating JS: %v", err)
		return err
	}

	log.Println(result)
	b.findElementAndScrollIntoView(ctx, "new to the group")
	return nil
}

func (b *Bot) findElementAndScrollIntoView(ctx context.Context, text string) error {
	var result string
	jsCode := fmt.Sprintf(`
		(function() {
			// Convert target text to lower case
			let targetText = "%s".toLowerCase();
			
			// Find all elements (e.g., span, div, etc.)
			let elements = document.body.querySelectorAll('span');
			let scrolled = false;

			for (let element of elements) {
				if (element.textContent.toLowerCase().includes(targetText)) {
					element.scrollIntoView({ behavior: 'smooth', block: 'center' });
					scrolled = true;
					// console.log('Element scrolled into view');
					return 'new group members scrolled into view';
				}
			}

			if (!scrolled) {
				console.log('No matching element found');
				return 'No matching element found';
			}
		})();
	`, text)

	// Evaluate the JavaScript code
	b.pause(10)
	err := chromedp.Evaluate(jsCode, &result).Do(ctx)
	if err != nil {
		log.Printf("Error evaluating JS: %v", err)
		return err
	}

	log.Println(result)
	return nil
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
