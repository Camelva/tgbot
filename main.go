package main

import (
	"fmt"
	"github.com/camelva/soundcloader"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
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
	config := loadConfigs("env.yml", "config.yml")

	bot, err := tgbotapi.NewBotAPI(config.Telegram.Token)
	if err != nil {
		logrus.WithError(err).Fatal("cant init telegram bot")
	}

	ttl, err := time.ParseDuration(config.Settings.FileTTL)
	if err != nil {
		ttl = time.Hour
	}

	c, err := NewClient(ClientConfig{
		b:              bot,
		dict:           getResponses(),
		ownerID:        config.Telegram.OwnerID,
		loadersLimit:   config.Settings.LoadersLimit,
		fileExpiration: ttl,
	})

	if err != nil {
		logrus.WithError(err).Fatal("cant create client")
	}

	//c.SetDebug(true)

	c.log.Info("Authorized on account ", bot.Self.UserName)

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
	c.log.Info("shutting down and sending logs..")
	if _, err := c.bot.Send(tgbotapi.NewDocumentUpload(int64(c.ownerID), c.logFile)); err != nil {
		logrus.WithError(err).Error("can't send message with logs to owner")
	}
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
