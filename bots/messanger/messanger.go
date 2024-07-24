package messanger

import (
	"context"
	// "fmt"
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

}
func (b *Bot) Date() string {
	t := time.Now()
	return t.Format("02 January 2006")
}

// pauses spider for given duration
func (b *Bot) Pause(second int) {
	time.Sleep(time.Duration(second) * time.Second)
}

// closes chromedp broswer instance
func (b *Bot) quit() {
	log.Println(b.Name, "done.")

	b.Quit()
}
func (b *Bot) Error(err error) {
	if err != nil {
		log.Println("*************************************")
		log.Println(b.Name, " Error:")
		log.Println(err.Error())
		log.Println(b.Name, " please restart bot")
		log.Println("*************************************")
		log.Fatal(err)

	}
}
