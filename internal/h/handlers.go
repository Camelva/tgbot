package h

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"go.uber.org/zap"
	"tgbot/internal/mux"
	"tgbot/internal/resp"
	"tgbot/internal/scloud"
)

func LogMessage(s *mux.Sender, ctx *ext.Context) error {
	// ignore empty messages, media or stickers
	if ctx.EffectiveMessage.Text == "" {
		return ext.EndGroups
	}

	s.Logger.Info("Got new message",
		zap.Int64("messageID", ctx.EffectiveMessage.MessageId),
		zap.Int64("userID", mux.GetUserID(ctx.EffectiveMessage)),
		zap.Int64("chatID", ctx.EffectiveChat.Id),
	)
	return ext.ContinueGroups
}

func LogCommand(s *mux.Sender, ctx *ext.Context) error {
	s.Logger.Info("It's command",
		zap.Int64("messageID", ctx.EffectiveMessage.MessageId),
		zap.String("value", ctx.EffectiveMessage.Text),
	)

	if err := s.Stats.Report(ctx.EffectiveMessage, false); err != nil {
		s.Logger.Error("can't send report", zap.Error(err))
	}

	return ext.ContinueGroups
}

func Start(s *mux.Sender, ctx *ext.Context) error {
	_, err := s.ReplyToMessage(
		s.AppCtx,
		ctx.EffectiveMessage,
		s.Resp.Get(tr.CmdStart, tr.GetLang(ctx.EffectiveMessage)),
		nil,
	)
	if err != nil {
		s.Logger.Error("failed to send Start command", zap.Error(err))
	}

	return ext.EndGroups
}

func Help(s *mux.Sender, ctx *ext.Context) error {
	_, err := s.ReplyToMessage(
		s.AppCtx,
		ctx.EffectiveMessage,
		s.Resp.Get(tr.CmdHelp, tr.GetLang(ctx.EffectiveMessage)),
		nil,
	)
	if err != nil {
		s.Logger.Error("failed to send Help command", zap.Error(err))
	}

	return ext.EndGroups
}

func Get(s *mux.Sender, ctx *ext.Context) error {
	_, err := s.ReplyToMessage(
		s.AppCtx,
		ctx.EffectiveMessage,
		s.Resp.Get(tr.CmdGet, tr.GetLang(ctx.EffectiveMessage)),
		&gotgbot.SendMessageOpts{ReplyMarkup: gotgbot.ForceReply{ForceReply: true, Selective: true}},
	)
	if err != nil {
		s.Logger.Error("failed to send Get command", zap.Error(err))
	}

	return ext.EndGroups
}

func Donate(s *mux.Sender, ctx *ext.Context) error {
	_, err := s.ReplyToMessage(
		s.AppCtx,
		ctx.EffectiveMessage,
		s.Resp.Get(tr.CmdDonate, tr.GetLang(ctx.EffectiveMessage)),
		nil,
	)
	if err != nil {
		s.Logger.Error("failed to send Donate command", zap.Error(err))
	}

	return ext.EndGroups
}

func SendLogs(s *mux.Sender, ctx *ext.Context) error {
	if !s.IsOwner(ctx.EffectiveMessage) {
		return ext.ContinueGroups
	}

	if err := s.ReportLogs(); err != nil {
		s.Logger.Error("can't send logs", zap.Error(err))
	}
	return ext.EndGroups
}

func Undefined(s *mux.Sender, ctx *ext.Context) error {
	// dont send error outside of private chat
	if ctx.EffectiveChat.Type != "private" {
		return ext.EndGroups
	}

	_, err := s.ReplyToMessage(
		s.AppCtx,
		ctx.EffectiveMessage,
		s.Resp.Get(tr.CmdUndefined, tr.GetLang(ctx.EffectiveMessage)),
		nil,
	)
	if err != nil {
		s.Logger.Error("failed to send Undefined command", zap.Error(err))
	}

	return ext.EndGroups
}

func ProcessURL(s *mux.Sender, ctx *ext.Context) error {
	success := scloud.ProcessURL(s, ctx)
	if err := s.Stats.Report(ctx.EffectiveMessage, success); err != nil {
		s.Logger.Error("can't send report", zap.Error(err))
	}
	return ext.EndGroups
}

func Default(s *mux.Sender, ctx *ext.Context) error {
	// dont send error outside of private chat
	if ctx.EffectiveChat.Type != "private" {
		return ext.EndGroups
	}

	_, err := s.ReplyToMessage(
		s.AppCtx,
		ctx.EffectiveMessage,
		s.Resp.Get(tr.ErrNotURL, tr.GetLang(ctx.EffectiveMessage)),
		nil,
	)
	if err != nil {
		s.Logger.Error("failed to send Undefined command", zap.Error(err))
	}

	if err := s.Stats.Report(ctx.EffectiveMessage, false); err != nil {
		s.Logger.Error("can't send report", zap.Error(err))
	}

	return ext.EndGroups
}
