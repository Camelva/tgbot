package main

import (
	"fmt"
	"github.com/camelva/soundcloader"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"
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
		log.Fatalf("cant init telegram bot: %s", err)
	}

	c, err := NewClient(ClientConfig{
		b:              bot,
		dict:           getResponses(),
		ownerID:        config.Telegram.OwnerID,
		loadersLimit:   10,
		fileExpiration: time.Hour,
	})

	if err != nil {
		log.Fatalf("cant create client: %s", err)
	}

	//c.SetDebug(true)

	log.Printf("Authorized on account %s\n", bot.Self.UserName)

	signal.Notify(c.shutdown, os.Interrupt, os.Kill)
	go func() {
		<-c.shutdown
		c.exit()
	}()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := c.bot.GetUpdatesChan(u)

	go c.downloader()

	for update := range updates {
		if update.Message != nil {
			go c.processMessage(update.Message)
		} else if update.ChannelPost != nil {
			go c.processMessage(update.ChannelPost)
		} else {
			continue
		}
	}

	<-c.done
	log.Println("shutting down..")
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
