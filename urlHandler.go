package main

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/camelva/soundcloader"
	"net/url"
	"tgbot/resp"
	"tgbot/telemetry"
	"unicode/utf16"
)

func checkURL(b *gotgbot.Bot, ctx *ext.Context) error {
	var success = false

	defer func() {
		_ = telemetry.SendReport(ctx, success)
	}()

	var urlsFromMsg []url.URL

	if len(ctx.EffectiveMessage.Entities) > 0 {
		for _, ent := range ctx.EffectiveMessage.Entities {
			if ent.Type != "url" {
				continue
			}
			// weird workaround because telegram count symbols in utf16
			text := utf16.Encode([]rune(ctx.EffectiveMessage.Text))
			link := string(utf16.Decode(text[ent.Offset:][:ent.Length]))

			u, err := url.Parse(link)
			if err != nil {
				continue
			}
			urlsFromMsg = append(urlsFromMsg, *u)
		}
	}

	if len(urlsFromMsg) < 1 {
		return replyNotURL(b, ctx)
	}

	var scURL *soundcloader.URLInfo
	for _, u := range urlsFromMsg {
		urlObj := soundcloader.ParseURL(&u)
		if urlObj != nil {
			scURL = urlObj
			break
		}
	}

	if scURL == nil {
		return replyNotSCURL(b, ctx)
	}

	// tell user we got their message
	tmpMsg, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ProcessStart, ctx.EffectiveUser.LanguageCode),
		&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
	if err != nil {
		return err
	}

	ctx.Data["tempMessage"] = tmpMsg
	ctx.Data["parsedURL"] = scURL

	err = loadSong(b, ctx)
	if err == nil {
		success = true
	}
	return err
}
