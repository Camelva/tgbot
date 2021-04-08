package main

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"tgbot/resp"
	"tgbot/telemetry"
)

func cmdStart(b *gotgbot.Bot, ctx *ext.Context) error {
	_ = telemetry.SendReport(ctx, false)
	log.WithField("command", "start").Info("got new command")
	_, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.CmdStart, ctx.EffectiveUser.LanguageCode),
		&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
	return err
}

func cmdHelp(b *gotgbot.Bot, ctx *ext.Context) error {
	_ = telemetry.SendReport(ctx, false)
	log.WithField("command", "help").Info("got new command")
	_, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.CmdHelp, ctx.EffectiveUser.LanguageCode),
		&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
	return err
}

func cmdUndefined(b *gotgbot.Bot, ctx *ext.Context) error {
	_ = telemetry.SendReport(ctx, false)
	log.WithField("command", ctx.EffectiveMessage.Text).Info("got undefined command")
	_, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.CmdUndefined, ctx.EffectiveUser.LanguageCode),
		&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
	return err
}

func replyNotURL(b *gotgbot.Bot, ctx *ext.Context) error {
	_ = telemetry.SendReport(ctx, false)
	// don't log people' text messages
	log.Info("got message without url")
	_, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrNotURL, ctx.EffectiveUser.LanguageCode),
		&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
	return err
}

func replyNotSCURL(b *gotgbot.Bot, ctx *ext.Context) error {
	_ = telemetry.SendReport(ctx, false)
	log.WithField("message", ctx.EffectiveMessage.Text).Info("got message without soundcloud url")
	_, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrNotSCURL, ctx.EffectiveUser.LanguageCode),
		&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
	return err
}

// echo replies to a messages with its own contents
//func echo(b *gotgbot.Bot, ctx *ext.Context) error {
//	_, err := ctx.EffectiveMessage.Reply(b, ctx.EffectiveMessage.Text, nil)
//	return err
//}
