package main

import (
	"fmt"
	"github.com/camelva/soundcloader"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
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

	//users map[int64]chan result
	cache map[int]*fileInfo
	//capacitor    chan struct{}

	capacitor *Capacitor

	//results chan result
	results *results

	shutdown chan os.Signal
	done     chan bool

	debug   bool
	log     *logrus.Logger
	logFile string
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

	logFile := fmt.Sprintf("%s.log", time.Now().Format("2006-01-02 15-04"))

	l := logrus.New()
	l.Hooks.Add(lfshook.NewHook(logFile, &logrus.JSONFormatter{}))

	l.SetLevel(logrus.TraceLevel)

	c := Client{
		bot:            conf.b,
		log:            l,
		logFile:        logFile,
		resp:           conf.dict,
		loader:         conf.scdl,
		ownerID:        conf.ownerID,
		fileExpiration: conf.fileExpiration,
	}

	//c.usersLoading = make(map[int64]*sync.WaitGroup)
	//c.users = make(map[int64]chan result)

	c.capacitor = &Capacitor{
		cond:      *sync.NewCond(&sync.Mutex{}),
		container: make(map[int]string),
		limit:     conf.loadersLimit,
	}

	c.cache = make(map[int]*fileInfo)
	//c.results = make(chan result, 100)
	c.results = NewResults(30, time.Second*15)

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

func deleteMessage(tgBot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	msgToDelete := tgbotapi.NewDeleteMessage(msg.Chat.ID, msg.MessageID)

	_, _ = tgBot.Send(msgToDelete)
}

func send(tgBot *tgbotapi.BotAPI, obj tgbotapi.Chattable) (message *tgbotapi.Message) {
	for range make([]int, 2) {
		sentMsg, err := tgBot.Send(obj)
		if err != nil {
			continue
		}
		return &sentMsg
	}
	return
}

// by default reply is true
func sendMessage(tgBot *tgbotapi.BotAPI, origin *tgbotapi.Message, text string, reply ...bool) *tgbotapi.Message {
	msgObj := tgbotapi.NewMessage(origin.Chat.ID, text)
	msgObj.ReplyToMessageID = origin.MessageID

	if len(reply) > 0 {
		if reply[0] == false {
			msgObj.ReplyToMessageID = 0
		}
	}

	msgObj.ParseMode = "HTML"
	return send(tgBot, msgObj)
}

func editMessage(tgBot *tgbotapi.BotAPI, msg *tgbotapi.Message, text string) *tgbotapi.Message {
	msgObj := tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, text)
	return send(tgBot, msgObj)
}

func replyNotURL(c *Client, msg *tgbotapi.Message) {
	c.log.WithField("value", msg.Text).Trace("no url here, exiting")

	// report to user only if its private chat
	if msg.Chat.IsPrivate() {
		sendMessage(c.bot, msg, getDict(c.resp, msg).MustLocalize(errNotURL))
	}
	return
}

func replyNotSC(c *Client, msg *tgbotapi.Message) {
	c.log.WithField("value", msg.Text).Trace("not soundcloud URL")

	// report to user only if its private chat
	if msg.Chat.IsPrivate() {
		sendMessage(c.bot, msg, getDict(c.resp, msg).MustLocalize(errNotSCURL))
	}
	return
}

func getDict(bundle *i18n.Bundle, msg *tgbotapi.Message) *i18n.Localizer {
	if msg != nil {
		if msg.From != nil {
			if msg.From.LanguageCode != "" {
				return i18n.NewLocalizer(bundle, msg.From.LanguageCode)
			}
		}
	}
	return i18n.NewLocalizer(bundle, "")
}

func (c *Client) loadersInfo() string {
	//keys := make([]int, 0, len(c.workingLoaders))
	//for k := range c.workingLoaders {
	//	keys = append(keys, k)
	//}
	return fmt.Sprintf(
		"\n#########\n# Active loaders amount: %d\n# Cached songs: %v\n# Limit: %d\n# Loading now: %s\n#########\n",
		c.capacitor.Len(), len(c.cache), c.capacitor.Max(), c.capacitor)
}

func (c *Client) exit() {
	c.bot.StopReceivingUpdates()
	c.log.Info("bot was turned off, finishing work..")
	for {
		if c.capacitor.Len() > 0 || c.results.Len() > 0 {
			c.log.Warn(fmt.Sprintf("%d songs left. %d user's loader(-s) still working. Songs: %s", c.capacitor.Len(), c.results.Len(), c.capacitor.String()))
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

func (c *Client) sendLogs() error {
	if _, err := c.bot.Send(tgbotapi.NewDocumentUpload(int64(c.ownerID), c.logFile)); err != nil {
		logrus.WithError(err).Error("can't send message with logs to owner")
		return err
	}
	return nil
}

func removeWhenExpire(fi *fileInfo, songLog *logrus.Entry) {
	songLog.Debug("scheduling file remove after %s", fi.ttl.String())

	time.AfterFunc(fi.ttl, fi.remove)
}

func (fi *fileInfo) remove() {
	_ = os.Remove(fi.fileLocation)
	delete(*fi.cacheContainer, fi.id)
}
