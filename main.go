package main

import (
	"fmt"
	"github.com/camelva/soundcloader"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"
)

type result struct {
	userID int64
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

	//go c.downloader()

	for update := range updates {
		if update.Message != nil {
			go processMessage(c, update.Message)
		} else if update.ChannelPost != nil {
			go processMessage(c, update.ChannelPost)
		} else {
			continue
		}
	}

	<-c.done
	c.log.Info("shutting down and sending logs..")
	_ = c.sendLogs()
}

type Capacitor struct {
	cond      sync.Cond
	container map[int]string
	limit     int
}

func (ca *Capacitor) Add(id int, permalink string) {
	ca.cond.L.Lock()
	defer ca.cond.L.Unlock()

	for len(ca.container) >= ca.limit {
		ca.cond.Wait()
	}

	ca.container[id] = permalink
}

func (ca *Capacitor) Remove(id int) {
	ca.cond.L.Lock()
	defer ca.cond.L.Unlock()

	if len(ca.container) >= ca.limit {
		// means Add() probably waiting, let them know
		ca.cond.Signal()
	}
	delete(ca.container, id)
}

func (ca *Capacitor) Len() int {
	return len(ca.container)
}

func (ca *Capacitor) Max() int {
	return ca.limit
}

func (ca *Capacitor) String() string {
	return fmt.Sprint(ca.container)
}

func memoryStats() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return fmt.Sprintf("\n########\n# Alloc = %v\n"+
		"# TotalAlloc = %v\n"+
		"# Sys = %v\n"+
		"# NumGC = %v\n########\n",
		m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.NumGC)
}
