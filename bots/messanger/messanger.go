package messanger

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"log"
	"sync"
	"time"
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

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false), // set headless to true for production
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.163 Safari/537.36"),
		chromedp.WindowSize(768, 1024), // Tablet size
	)
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	// defer cancel() // Removed to keep the browser open

	ctx, cancel = chromedp.NewContext(ctx)
	b.Quit = cancel

	log.Println(b.Name, " loading url: ", b.URL)

	err := chromedp.Run(ctx,
		chromedp.Navigate(b.URL),
	)
	b.error(err)

	// Keep the browser open
	b.waitForUserInput()
}

func (b *Bot) waitForUserInput() {
	log.Println("Press Enter to exit...")
	var input string
	fmt.Scanln(&input)
	b.quit()
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
