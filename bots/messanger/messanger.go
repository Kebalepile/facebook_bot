package messanger

import (
	"bufio"
	"context"
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
	log.Println("init: ", b.Name)

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

	log.Println(b.Name, " loading url: ", b.URL)

	err := chromedp.Run(ctx,
		chromedp.Navigate(b.URL),
		b.waitVisibleAndSendKeys(`//*[@placeholder='Email address or phone number']`, email),
		b.waitVisibleAndSendKeys(`//*[@placeholder='Password']`, password),
		b.waitVisibleAndClick(`//button[text()='Log in']`),
	)
	b.error(err)

	// Keep the browser open and wait for user input
	b.waitForUserInput()
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
			log.Println("Error reading input:", err)
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
	log.Println(b.Name, "done.")
	b.Quit()
}

func (b *Bot) error(err error) {
	if err != nil {
		log.Println("*************************************")
		log.Println(b.Name, " Error:")
		log.Println(err.Error())
		log.Println(b.Name, " please restart bot")
		log.Println("*************************************")
		log.Fatal(err)
	}
}
