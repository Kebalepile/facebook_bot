package main

import (
	"github.com/Kebalepile/facebook_bot/bots/messanger"
	"github.com/Kebalepile/facebook_bot/hooks/types"
	"log"
	"sync"
)

func main() {

	log.Println("Init facebook Bots")
	mBot := messanger.Bot{
		Name:    "Messanger Bot",
		URL:     "https://www.facebook.com/",
		Message: "hello there, ke nna john",
	}

	bots := []types.FacebookBot{
		&mBot,
	}

	var wg sync.WaitGroup
	for _, bot := range bots {
		wg.Add(1)
		go bot.Run(&wg)
	}

	wg.Wait()

}
