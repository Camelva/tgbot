package main

import (
	"fmt"
	"github.com/camelva/soundcloader"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"net/url"
	"runtime"
	"strconv"
	"unicode/utf16"
)

func processMessage(c *Client, msg *tgbotapi.Message) {
	// if its sticker or so on
	if msg.Text == "" {
		return
	}

	var uid int64
	if msg.From != nil {
		uid = int64(msg.From.ID)
	} else {
		uid = msg.Chat.ID
	}

	c.log.WithField("value", msg.Text).WithField("userID", uid).Info("Got new message")
	reportMessage(msg, c.log)

	if msg.IsCommand() {
		processCmd(c, msg)
		return
	}

	//urlObj := c.loader.Parse(msg.Text)
	var urlsFromMsg []url.URL

	if len(msg.Entities) > 0 {
		for _, ent := range msg.Entities {
			if !ent.IsURL() {
				continue
			}
			// weird workaround because telegram count symbols in utf16
			text := utf16.Encode([]rune(msg.Text))
			m := string(utf16.Decode(text[ent.Offset:][:ent.Length]))
			u, err := url.Parse(m)
			if err != nil {
				continue
			}
			urlsFromMsg = append(urlsFromMsg, *u)
		}
	}

	if len(urlsFromMsg) < 1 {
		replyNotURL(c, msg)
		return
	}

	var scURL *soundcloader.URLInfo
	for _, u := range urlsFromMsg {
		urlObj := c.loader.ParseURL(&u)
		if urlObj != nil {
			scURL = urlObj
			break
		}
	}

	if scURL == nil {
		replyNotSC(c, msg)
		return
	}

	// tell user we got their message
	tmpMsg := sendMessage(c.bot, msg, getDict(c.resp, msg).MustLocalize(processStart))

	song, err := c.loader.GetURL(scURL)
	if err != nil {
		scURLLog := c.log.WithField("value", scURL.String())
		if err == soundcloader.NotSong {
			scURLLog.Info("not song url, exiting..")
			editMessage(c.bot, tmpMsg, getDict(c.resp, msg).MustLocalize(errUnsupportedFormat))
		} else {
			scURLLog.WithError(err).Error("can't get url")
			editMessage(c.bot, tmpMsg, getDict(c.resp, msg).MustLocalize(errUnavailableSong))
		}
		return
	}
	r := result{userID: uid, msg: *msg, tmpMsg: *tmpMsg, song: *song}
	//c.results <- result{userID: uid, msg: *msg, tmpMsg: *tmpMsg, song: *song}
	c.results.Add(c, r)
	return
}

func processCmd(c *Client, msg *tgbotapi.Message) {
	c.log.WithField("value", msg.Command()).Trace("its command, responding..")
	if msg.Command() == "help" {
		sendMessage(c.bot, msg, getDict(c.resp, msg).MustLocalize(cmdHelp))
		return
	}

	if msg.Command() == "start" {
		sendMessage(c.bot, msg, getDict(c.resp, msg).MustLocalize(cmdStart))
		return
	}

	if adminCommand(c, msg) {
		return
	}

	sendMessage(c.bot, msg, getDict(c.resp, msg).MustLocalize(cmdUndefined))
	return
}

func adminCommand(c *Client, msg *tgbotapi.Message) (ok bool) {
	if !c.sentByOwner(msg) {
		return false
	}
	ok = true

	switch msg.Command() {
	case "debug":
		d := msg.CommandArguments()
		if d == "true" || d == "1" || d == "yes" {
			c.SetDebug(true)
		} else if d == "false" || d == "0" || d == "no" {
			c.SetDebug(false)
		}
		sendMessage(c.bot, msg, fmt.Sprintf("Debug = %t", c.debug))
		return
	case "stats":
		sendMessage(c.bot, msg, memoryStats())
		return
	case "queue":
		sendMessage(c.bot, msg, c.loadersInfo())
		return
	case "gr":
		sendMessage(c.bot, msg, fmt.Sprintf("Goroutines number: %d", runtime.NumGoroutine()))
		return
	case "setTTL":
		i := msg.CommandArguments()
		if err := changeConfig(Settings{FileTTL: i}); err != nil {
			sendMessage(c.bot, msg, fmt.Sprintf("can't apply changes: %s", err))
			return
		}
		sendMessage(c.bot, msg, "TTL changed!")
		return
	case "setLimit":
		i, err := strconv.Atoi(msg.CommandArguments())
		if err != nil {
			i = 10
		}
		if err := changeConfig(Settings{LoadersLimit: i}); err != nil {
			sendMessage(c.bot, msg, fmt.Sprintf("can't apply changes: %s", err))
			return
		}
		sendMessage(c.bot, msg, "Limit changed!")
		return
	case "logs":
		if err := c.sendLogs(); err != nil {
			sendMessage(c.bot, msg, fmt.Sprintf("can't send logs: %v", err))
		}
		return
	case "stop":
		c.exit()
		sendMessage(c.bot, msg, "Done, wait for logs file")
		if err := c.sendLogs(); err != nil {
			sendMessage(c.bot, msg, fmt.Sprintf("can't send logs: %v", err))
		}
		return
	}
	return false
}
