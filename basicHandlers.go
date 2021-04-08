package main

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"tgbot/resp"
	"tgbot/telemetry"
)

func logMessage(b *gotgbot.Bot, ctx *ext.Context) error {
	// sticker or anything non-text
	if ctx.EffectiveMessage.Text != "" {
		log.
			WithField("messageID", ctx.EffectiveMessage.MessageId).
			WithField("userID", ctx.EffectiveUser.Id).
			Info("Got new message")
	}
	return ext.ContinueGroups
}

func logCmd(b *gotgbot.Bot, ctx *ext.Context) error {
	log.
		WithField("messageID", ctx.EffectiveMessage.MessageId).
		WithField("value", ctx.EffectiveMessage.Text).
		Info("it's command, responding..")

	_ = telemetry.SendReport(ctx, false)

	return ext.ContinueGroups
}

func cmdStart(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.CmdStart, ctx.EffectiveUser.LanguageCode),
		&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})

	if err != nil {
		log.WithError(err).Error("while responding")
	}

	return ext.EndGroups
}

func cmdHelp(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.CmdHelp, ctx.EffectiveUser.LanguageCode),
		&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})

	if err != nil {
		log.WithError(err).Error("while responding")
	}

	return ext.EndGroups
}

func cmdUndefined(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.CmdUndefined, ctx.EffectiveUser.LanguageCode),
		&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})

	if err != nil {
		log.WithError(err).Error("while responding")
	}

	return ext.EndGroups
}

func replyNotURL(b *gotgbot.Bot, ctx *ext.Context) error {
	log.WithField("messageID", ctx.EffectiveMessage.MessageId).Info("no url here, exiting")

	_ = telemetry.SendReport(ctx, false)

	if ctx.EffectiveChat.Type != "private" {
		return nil
	}

	_, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrNotURL, ctx.EffectiveUser.LanguageCode),
		&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})

	if err != nil {
		log.WithError(err).Error("while responding")
	}

	return ext.EndGroups
}

func replyNotSCURL(b *gotgbot.Bot, ctx *ext.Context) error {
	log.
		WithField("messageID", ctx.EffectiveMessage.MessageId).
		WithField("value", ctx.EffectiveMessage.Text).
		Info("not soundcloud url")

	_ = telemetry.SendReport(ctx, false)

	_, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrNotSCURL, ctx.EffectiveUser.LanguageCode),
		&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})

	if err != nil {
		log.WithError(err).Error("while responding")
	}

	return ext.EndGroups
}

// echo replies to a messages with its own contents
//func echo(b *gotgbot.Bot, ctx *ext.Context) error {
//	_, err := ctx.EffectiveMessage.Reply(b, ctx.EffectiveMessage.Text, nil)
//	return err
//}
