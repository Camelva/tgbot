package main

import (
	"fmt"
	"github.com/camelva/soundcloader"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/url"
	"runtime"
	"unicode/utf16"
)

func (c *Client) processMessage(msg *tgbotapi.Message) {
	log.Println("Got new message")
	reportMessage(msg)

	if msg.IsCommand() {
		c.processCmd(msg)
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
		c.returnNoURL(msg)
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
		c.returnNoURL(msg)
		return
	}

	// tell user we got their message
	tmpMsg := c.sendMessage(msg, c.getDict(msg).MustLocalize(processStart), true)

	song, err := c.loader.GetURL(scURL)
	if err != nil {
		if err == soundcloader.NotSong {
			log.Println("not song url, exiting..")
			c.editMessage(msg, c.getDict(msg).MustLocalize(errUnsupportedFormat))
		} else {
			log.Printf("can't get url: %s | Error: %s", scURL.String(), err)
			c.editMessage(tmpMsg, c.getDict(msg).MustLocalize(errUnavailableSong))
		}
		return
	}

	c.results <- result{*msg, *tmpMsg, *song}
	return
}

func (c *Client) processCmd(msg *tgbotapi.Message) {
	log.Println("its command, responding..")
	if msg.Command() == "help" {
		c.sendMessage(msg, c.getDict(msg).MustLocalize(cmdHelp), true)
		return
	}

	if msg.Command() == "start" {
		c.sendMessage(msg, c.getDict(msg).MustLocalize(cmdStart), true)
		return
	}

	if c.adminCommand(msg) {
		return
	}

	c.sendMessage(msg, c.getDict(msg).MustLocalize(cmdUndefined), true)
	return
}

func (c *Client) adminCommand(msg *tgbotapi.Message) (ok bool) {
	if !c.sentByOwner(msg) {
		return false
	}

	switch msg.Command() {
	case "debug":
		d := msg.CommandArguments()
		if d == "true" || d == "1" || d == "yes" {
			c.SetDebug(true)
		} else if d == "false" || d == "0" || d == "no" {
			c.SetDebug(false)
		}
		c.sendMessage(msg, fmt.Sprintf("Debug = %t", c.debug), true)
		return true
	case "stats":
		c.sendMessage(msg, getUsageStats(), true)
		return true
	case "queue":
		c.sendMessage(msg, c.loadersInfo(), true)
		return true
	case "gr":
		c.sendMessage(msg, fmt.Sprintf("Goroutines number: %d", runtime.NumGoroutine()), true)
		return true
	//case "memorystats":
	//	c.sendMessage(msg, "hey", true)
	//	return true
	case "stop":
		c.exit()
		c.sendMessage(msg, "Done!", true)
		return true
	}
	return false
}
