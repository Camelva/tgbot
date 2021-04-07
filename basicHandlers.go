package main

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"tgbot/resp"
)

func cmdStart(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.CmdStart, ctx.EffectiveUser.LanguageCode),
		&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
	return err
}

func cmdHelp(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.CmdHelp, ctx.EffectiveUser.LanguageCode),
		&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
	return err
}

func cmdUndefined(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.CmdUndefined, ctx.EffectiveUser.LanguageCode),
		&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
	return err
}

func replyNotURL(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrNotURL, ctx.EffectiveUser.LanguageCode),
		&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
	return err
}

func replyNotSCURL(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrNotSCURL, ctx.EffectiveUser.LanguageCode),
		&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
	return err
}

// echo replies to a messages with its own contents
//func echo(b *gotgbot.Bot, ctx *ext.Context) error {
//	_, err := ctx.EffectiveMessage.Reply(b, ctx.EffectiveMessage.Text, nil)
//	return err
//}
