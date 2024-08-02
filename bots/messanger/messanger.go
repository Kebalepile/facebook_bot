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

func (b *Bot) endUserMessage() (string, error) {
	log.Printf("%v: Enter text to be sent via messenger chat", b.Name)
	reader := bufio.NewReader(os.Stdin)

	input, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading input:", err)
		return "", err
	}
	input = strings.TrimSpace(input)
	if len(input) > 0 {
		return input, nil
	} else {
		log.Println("Invalid input. Type your message and press Enter to proceed...")
		return b.endUserMessage()
	}

}

func (b *Bot) navigate_to_messenger(ctx context.Context) {
	messengerSvg := "m459.603 1077.948-1.762 2.851a.89.89 0 0 1-1.302.245l-1.402-1.072a.354.354 0 0 0-.433.001l-1.893 1.465c-.253.196-.583-.112-.414-.386l1.763-2.851a.89.89 0 0 1 1.301-.245l1.402 1.072a.354.354 0 0 0 .434-.001l1.893-1.465c.253-.196.582.112.413.386M456 1073.5c-3.38 0-6 2.476-6 5.82 0 1.75.717 3.26 1.884 4.305.099.087.158.21.162.342l.032 1.067a.48.48 0 0 0 .674.425l1.191-.526a.473.473 0 0 1 .32-.024c.548.151 1.13.231 1.737.231 3.38 0 6-2.476 6-5.82 0-3.344-2.62-5.82-6-5.82"

	b.pause(10)
	log.Println("Navigating to messenger")

	textMessage, err := b.endUserMessage()
	b.error(err)
	jsCode := fmt.Sprintf(`
		function clickTargetElement() {
			const targetElements = document.querySelectorAll('.x1i10hfl.x1qjc9v5.xjbqb8w.xjqpnuy.xa49m3k.xqeqjp1.x2hbi6w.x13fuv20.xu3j5b3.x1q0q8m5.x26u7qi.x972fbf.xcfux6l.x1qhh985.xm0m39n.x9f619.x1ypdohk.xdl72j9.x2lah0s.xe8uvvx.xdj266r.x11i5rnm.xat24cr.x1mh8g0r.x2lwn1j.xeuugli.xexx8yu.x4uap5.x18d9i69.xkhd6sd.x1n2onr6.x16tdsg8.x1hl2dhg.xggy1nq.x1ja2u2z.x1t137rt.x1o1ewxj.x3x9cwd.x1e5q0jg.x13rtm0m.x1q0g3np.x87ps6o.x1lku1pv.x1a2a7pz.x1lliihq');
	
			if (targetElements.length) {
				targetElements[0].click();
				
					// const editableDiv = document.querySelector('div[aria-describedby^="Write to "][aria-label="Message"][contenteditable="true"]');
					// if (editableDiv) {
					// 	editableDiv.focus();
						
					// 	const span = document.createElement('span');
					// 	span.setAttribute('data-lexical-text', 'true');
					// 	span.innerText = %q;
						
					// 	editableDiv.appendChild(span);
						
					// 	const enterEvent = new KeyboardEvent('keydown', {
					// 		key: 'Enter',
					// 		keyCode: 13,
					// 		which: 13,
					// 		bubbles: true,
					// 		cancelable: true
					// 	});
					// 	editableDiv.dispatchEvent(enterEvent);	
					// }
			} else {
				console.log('Target element not found');
			}
		}
		clickTargetElement();
	`, textMessage)

	err = chromedp.Run(ctx,
		b.clickSvgParentElementByPath(messengerSvg),
		chromedp.Evaluate(jsCode, nil),
		b.clickSvgParentElementByPath("m98.095 917.155 7.75 7.75a.75.75 0 0 0 1.06-1.06l-7.75-7.75a.75.75 0 0 0-1.06 1.06z"),
	)
	b.error(err)

	b.waitForUserInput()
}

func (b *Bot) error(err error) {
	if err != nil {
		log.Fatal(err)
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
			jsCode := fmt.Sprintf(
				`(function() {
                function findGroupSection() {
                    return Array.from(document.querySelectorAll('span')).find(span =>
                    span.textContent.toLowerCase().includes('public') && (span.textContent.toLowerCase().includes('members') || span.textContent.toLowerCase().includes('people'))
                    );
                }

                let publicGroup = findGroupSection();

                if (!publicGroup) {
                    return 'Group is not public';
                }

                let targetText = %q.toLowerCase(); // This will be dynamic in your actual use case
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
            })();`, searchGroup)
			err := chromedp.Evaluate(jsCode, &result).Do(ctx)
			if err != nil {
				log.Printf("Error evaluating findGroupSection JS: %v", err)
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
func (b *Bot) addFriendsFromNewToGroupSection(ctx context.Context) chromedp.ActionFunc {

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
				return err
			}

			// Execute the JavaScript to add friends from the "New to the group" section
			var result map[string]interface{}
			err = chromedp.Evaluate(
				`
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
				log.Printf("Error executing addFriendsFromNewToGroupSection JavaScript: %v", err)
				return err
			}

			return nil
		})
	}
	return nil
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

// clickElementWithText clicks on the element containing the text 'People', case-insensitively
func (b *Bot) clickElementWithText(ctx context.Context, text string) error {
	var result string
	jsCode := fmt.Sprintf(
		`
		(function() {
			// Convert target text to lower case
			let targetText = %q.toLowerCase();
			
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
		log.Printf("Error evaluating clickElementWithText JS: %v", err)
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
			let targetText = %q.toLowerCase();
			
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
		log.Printf("Error evaluating findElementAndScrollIntoView JS: %v", err)
		return err
	}

	log.Println(result)
	return nil
}

// Wait for user to type 'continue' and press Enter
func (b *Bot) waitForContinue() {
	log.Printf("%v paused waiting for your command. \nType 'continue' and press Enter to proceed...", b.Name)

	reader := bufio.NewReader(os.Stdin)

	input, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading input:", err)
		return
	}

	// Trim any whitespace or newline characters from the input
	input = strings.TrimSpace(input)

	// Check if the input is "continue"
	if input == "continue" {
		return
	} else {
		log.Println("Invalid input. Type 'continue' and press Enter to proceed...")
		b.waitForContinue()
	}

}

// Pause execution for a specified number of seconds
func (b *Bot) pause(seconds int) {
	log.Printf("Pausing for %d seconds...", seconds)
	time.Sleep(time.Duration(seconds) * time.Second)
}

func (b *Bot) waitForUserInput() {
	log.Println("Type 'e' or 'exit' and press Enter to exit...")

	reader := bufio.NewReader(os.Stdin)

	input, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error reading input: %v", err)
		return

	}

	input = strings.TrimSpace(input)

	if input == "e" || input == "exit" {
		return
	} else {
		log.Println("Invalid input. Type 'e' or 'exit' and press Enter to exit...")
		b.waitForUserInput()
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
