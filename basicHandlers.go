package main

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"tgbot/resp"
	"tgbot/telemetry"
)

func logMessage(_ *gotgbot.Bot, ctx *ext.Context) error {
	// sticker or anything non-text
	if ctx.EffectiveMessage.Text != "" {
		log.
			WithField("messageID", ctx.EffectiveMessage.MessageId).
			WithField("userID", getUserID(ctx)).
			Info("Got new message")
	}
	return ext.ContinueGroups
}

func logCmd(_ *gotgbot.Bot, ctx *ext.Context) error {
	log.
		WithField("messageID", ctx.EffectiveMessage.MessageId).
		WithField("value", ctx.EffectiveMessage.Text).
		Info("it's command, responding..")

	_ = telemetry.SendReport(ctx, false)

	return ext.ContinueGroups
}

func cmdStart(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.CmdStart, getLang(ctx)),
		&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})

	if err != nil {
		log.WithError(err).Error("while responding")
	}

	return ext.EndGroups
}

func cmdHelp(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.CmdHelp, getLang(ctx)),
		&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})

	if err != nil {
		log.WithError(err).Error("while responding")
	}

	return ext.EndGroups
}

func cmdLogs(b *gotgbot.Bot, ctx *ext.Context) error {
	if !isOwner(getUserID(ctx)) {
		return cmdUndefined(b, ctx)
	}

	e := sendLogs(b)
	if e != nil {
		_, err := b.SendMessage(ctx.EffectiveChat.Id, "can't send logs", nil)

		if err != nil {
			log.WithError(err).Error("while responding")
		}
	}

	return ext.EndGroups
}

func cmdUndefined(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.CmdUndefined, getLang(ctx)),
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

	_, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrNotURL, getLang(ctx)),
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

	if ctx.EffectiveChat.Type != "private" {
		return nil
	}

	_, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrNotSCURL, getLang(ctx)),
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
