package main

import (
	"fmt"
	"github.com/camelva/soundcloader"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"runtime"
)

type result struct {
	msg    tgbotapi.Message
	tmpMsg tgbotapi.Message
	song   soundcloader.Song
}

func main() {
	config := loadConfig("config.yml")

	bot, err := tgbotapi.NewBotAPI(config.Telegram.Token)
	if err != nil {
		log.Fatal(err)
	}

	c := NewClient(bot, soundcloader.DefaultClient, getResponses())

	c.SetOwner(config.Telegram.OwnerID)

	c.bot.Debug = false

	log.Printf("Authorized on account %s\n", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := c.bot.GetUpdatesChan(u)

	resChan := make(chan result)

	go c.downloader(resChan)

	for update := range updates {
		if update.Message != nil {
			go c.processMessage(update.Message, resChan)
		} else if update.ChannelPost != nil {
			go c.processMessage(update.ChannelPost, resChan)
		} else {
			continue
		}
	}
}

func (c *Client) sentByOwner(msg *tgbotapi.Message) bool {
	if msg.From == nil {
		return false
	}
	if msg.From.ID != c.ownerID {
		return false
	}
	return true
}

func getUsageStats() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return fmt.Sprintf("\n########\n# Alloc = %v\n"+
		"# TotalAlloc = %v\n"+
		"# Sys = %v\n"+
		"# NumGC = %v\n########\n",
		m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.NumGC)
}
