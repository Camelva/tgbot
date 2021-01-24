package main

import (
	"fmt"
	"github.com/camelva/soundcloader"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"log"
	"os"
	"sync"
	"time"
)

type Client struct {
	bot    *tgbotapi.BotAPI
	resp   *i18n.Bundle
	loader *soundcloader.Client

	ownerID        int
	fileExpiration time.Duration

	loadersLimit int
	usersLoading map[int64]*sync.WaitGroup
	cache        map[int]*fileInfo
	capacitor    chan struct{}

	results chan result

	shutdown chan os.Signal
	done     chan bool

	debug bool
}

type ClientConfig struct {
	b    *tgbotapi.BotAPI
	scdl *soundcloader.Client
	dict *i18n.Bundle

	ownerID        int
	loadersLimit   int
	fileExpiration time.Duration
	debug          bool
}

func NewClient(conf ClientConfig) (*Client, error) {
	if conf.b == nil {
		return nil, fmt.Errorf("need bot instance")
	}
	if conf.dict == nil {
		return nil, fmt.Errorf("need translation bundle")
	}

	if conf.scdl == nil {
		conf.scdl = soundcloader.DefaultClient
	}
	if conf.loadersLimit == 0 {
		conf.loadersLimit = 10
	}
	if conf.fileExpiration == 0 {
		conf.fileExpiration = time.Hour
	}

	c := Client{
		bot:            conf.b,
		resp:           conf.dict,
		loader:         conf.scdl,
		ownerID:        conf.ownerID,
		loadersLimit:   conf.loadersLimit,
		fileExpiration: conf.fileExpiration,
	}

	c.usersLoading = make(map[int64]*sync.WaitGroup)
	c.capacitor = make(chan struct{}, c.loadersLimit)
	c.cache = make(map[int]*fileInfo)
	c.results = make(chan result)

	c.shutdown = make(chan os.Signal, 1)
	c.done = make(chan bool, 1)

	return &c, nil
}

func (c *Client) SetOwner(id int) {
	c.ownerID = id
}

func (c *Client) Debug() bool {
	return c.debug
}

func (c *Client) SetDebug(b bool) {
	c.bot.Debug = b
	c.loader.Debug = b
	c.debug = b
}

func (c *Client) deleteMessage(msg *tgbotapi.Message) {
	msgToDelete := tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID)

	_, _ = c.bot.Send(msgToDelete)
}

func (c *Client) send(obj tgbotapi.Chattable) (message *tgbotapi.Message) {
	for range make([]int, 2) {
		sentMsg, err := c.bot.Send(obj)
		if err != nil {
			continue
		}
		return &sentMsg
	}
	return
}

func (c *Client) sendMessage(receivedMsg *tgbotapi.Message, text string, reply bool) (message *tgbotapi.Message) {
	msgObj := tgbotapi.NewMessage(receivedMsg.Chat.ID, text)
	if reply {
		msgObj.ReplyToMessageID = receivedMsg.MessageID
	}
	msgObj.ParseMode = "HTML"
	return c.send(msgObj)
}

func (c *Client) editMessage(msg *tgbotapi.Message, text string) (message *tgbotapi.Message) {
	msgObj := tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, text)
	return c.send(msgObj)
}

func (c *Client) returnNoURL(msg *tgbotapi.Message) {
	log.Println("no url, exiting..")
	if msg.Chat.IsPrivate() {
		c.sendMessage(msg, c.getDict(msg).MustLocalize(errNotURL), true)
	}
	return
}

func (c *Client) returnNoSCURL(msg *tgbotapi.Message) {
	log.Println("no soundcloud url, exiting..")
	if msg.Chat.IsPrivate() {
		c.sendMessage(msg, c.getDict(msg).MustLocalize(errNotSCURL), true)
	}
	return
}

func (c *Client) getDict(msg *tgbotapi.Message) *i18n.Localizer {
	if msg != nil {
		if msg.From != nil {
			if msg.From.LanguageCode != "" {
				return i18n.NewLocalizer(c.resp, msg.From.LanguageCode)
			}
		}
	}
	return i18n.NewLocalizer(c.resp, "")
}

func (c *Client) loadersInfo() string {
	//keys := make([]int, 0, len(c.workingLoaders))
	//for k := range c.workingLoaders {
	//	keys = append(keys, k)
	//}
	return fmt.Sprintf(
		"\n#########\n# Active loaders amount: %d\n# Cached songs: %v\n# Limit: %d\n#########\n",
		len(c.capacitor), len(c.cache), c.loadersLimit)
}

func (c *Client) exit() {
	c.bot.StopReceivingUpdates()
	log.Println("bot was turned off, finishing work...")
	for {
		if len(c.capacitor) > 0 || len(c.results) > 0 {
			log.Printf("still have [%d] songs to download and [%d] messages to process, waiting..", len(c.capacitor), len(c.results))
			time.Sleep(time.Second * 30)
			continue
		}
		break
	}
	c.clearCache()
	c.done <- true
}

func (c *Client) clearCache() {
	for _, fi := range c.cache {
		fi.remove()
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

func (c *Client) removeWhenExpire(fi *fileInfo) {
	if c.debug {
		log.Printf("remove [%d] after %s", fi.id, fi.ttl)
	}
	time.AfterFunc(fi.ttl, fi.remove)
}

func (fi *fileInfo) remove() {
	_ = os.Remove(fi.fileLocation)
	delete(*fi.cacheContainer, fi.id)
}
