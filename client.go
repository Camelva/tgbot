package main

import (
	"fmt"
	"github.com/camelva/soundcloader"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"log"
)

type Client struct {
	bot    *tgbotapi.BotAPI
	resp   *i18n.Bundle
	loader *soundcloader.Client

	workingLoaders map[int]*fileState
	capacitor      chan struct{}

	ownerID int
	debug   bool
}

func NewClient(b *tgbotapi.BotAPI, sl *soundcloader.Client, resp *i18n.Bundle) *Client {
	return &Client{
		bot:    b,
		loader: sl,
		resp:   resp,
	}
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
	keys := make([]int, 0, len(c.workingLoaders))
	for k := range c.workingLoaders {
		keys = append(keys, k)
	}
	return fmt.Sprintf(
		"\n#########\n# Active loaders amount: %d\n# Loaders: %v\n# Limit: %d\n#########\n",
		len(c.capacitor), keys, cap(c.capacitor))
}
