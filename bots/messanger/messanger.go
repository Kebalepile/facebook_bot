package messanger

import (
	"bufio"
	"context"
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
	b.pause(10)
	b.waitForContinue()

	// Handle alerts
	b.navigate_to_messenger(ctx)
}

// Wait for user to type 'continue' and press Enter
func (b *Bot) waitForContinue() {
	log.Printf("%v paused waiting for your command. \nType 'continue' and press Enter to proceed...", b.Name)

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

// Pause execution for a specified number of seconds
func (b *Bot) pause(seconds int) {
	log.Printf("Pausing for %d seconds...", seconds)
	time.Sleep(time.Duration(seconds) * time.Second)
}

// Send a direct message
func (b *Bot) sendDirectMessage(ctx context.Context) chromedp.ActionFunc {
	b.pause(10)
	log.Println("sending direct message")

	return chromedp.ActionFunc(func(ctx context.Context) error {
		err := chromedp.Evaluate(`
		function clickTargetElement() {
			// Select the element with the specific class name
			const targetElements = document.querySelectorAll('.x1i10hfl.x1qjc9v5.xjbqb8w.xjqpnuy.xa49m3k.xqeqjp1.x2hbi6w.x13fuv20.xu3j5b3.x1q0q8m5.x26u7qi.x972fbf.xcfux6l.x1qhh985.xm0m39n.x9f619.x1ypdohk.xdl72j9.x2lah0s.xe8uvvx.xdj266r.x11i5rnm.xat24cr.x1mh8g0r.x2lwn1j.xeuugli.xexx8yu.x4uap5.x18d9i69.xkhd6sd.x1n2onr6.x16tdsg8.x1hl2dhg.xggy1nq.x1ja2u2z.x1t137rt.x1o1ewxj.x3x9cwd.x1e5q0jg.x13rtm0m.x1q0g3np.x87ps6o.x1lku1pv.x1a2a7pz.x1lliihq');

			// Check if the target element exists
			if (targetElements.length) {
				// targetElements.forEach(elem => elem.click());
				targetElements[0].click();
			} else {
				console.log('Target element not found');
			}
		}
		clickTargetElement();
		`, nil).Do(ctx)

		if err != nil {
			log.Printf("Error executing Send DM JavaScript: %v", err)
			return err
		}
		return nil
	})
}

func (b *Bot) navigate_to_messenger(ctx context.Context) {
	messengerSvg := "m459.603 1077.948-1.762 2.851a.89.89 0 0 1-1.302.245l-1.402-1.072a.354.354 0 0 0-.433.001l-1.893 1.465c-.253.196-.583-.112-.414-.386l1.763-2.851a.89.89 0 0 1 1.301-.245l1.402 1.072a.354.354 0 0 0 .434-.001l1.893-1.465c.253-.196.582.112.413.386M456 1073.5c-3.38 0-6 2.476-6 5.82 0 1.75.717 3.26 1.884 4.305.099.087.158.21.162.342l.032 1.067a.48.48 0 0 0 .674.425l1.191-.526a.473.473 0 0 1 .32-.024c.548.151 1.13.231 1.737.231 3.38 0 6-2.476 6-5.82 0-3.344-2.62-5.82-6-5.82"

	b.pause(10)
	log.Println("Navigating to messenger")

	err := chromedp.Run(ctx,
		b.clickSvgParentElementByPath(messengerSvg),
		b.sendDirectMessage(ctx),
	)
	b.error(err)
	b.waitForUserInput()
}

func (b *Bot) error(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Utility functions
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
						parent := node.Parent.Parent
						if parent != nil {
							return chromedp.Click(parent.FullXPath()).Do(ctx)
						}
					}
				}
			}
			return nil
		}),
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

		input = strings.TrimSpace(input)

		if input == "e" || input == "exit" {
			break
		} else {
			log.Println("Invalid input. Type 'e' or 'exit' and press Enter to exit...")
		}
	}
}

// Read .env variables to be used
func (b *Bot) env() map[string]string {
	variables, err := godotenv.Read()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	return variables
}
