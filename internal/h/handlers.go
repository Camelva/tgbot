package h

import (
	"context"
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
	)
	return ext.ContinueGroups
}

func LogCommand(s *mux.Sender, ctx *ext.Context) error {
	s.Logger.Info("It's command",
		zap.Int64("messageID", ctx.EffectiveMessage.MessageId),
		zap.String("value", ctx.EffectiveMessage.Text),
	)

	if err := s.Stats.Report(ctx.Message, false); err != nil {
		s.Logger.Error("can't send report", zap.Error(err))
	}

	return ext.ContinueGroups
}

func Start(s *mux.Sender, ctx *ext.Context) error {
	// let's hope ext.Context will be updated some day..
	cntxt := context.Background()

	_, err := s.ReplyToMessage(
		cntxt,
		ctx.EffectiveMessage,
		s.Resp.Get(tr.CmdStart, tr.GetLang(ctx.Message)),
		nil,
	)
	if err != nil {
		s.Logger.Error("failed to send Start command", zap.Error(err))
	}

	return ext.EndGroups
}

func Help(s *mux.Sender, ctx *ext.Context) error {
	// let's hope ext.Context will be updated some day..
	cntxt := context.Background()

	_, err := s.ReplyToMessage(
		cntxt,
		ctx.EffectiveMessage,
		s.Resp.Get(tr.CmdHelp, tr.GetLang(ctx.Message)),
		nil,
	)
	if err != nil {
		s.Logger.Error("failed to send Help command", zap.Error(err))
	}

	return ext.EndGroups
}

func SendLogs(s *mux.Sender, ctx *ext.Context) error {
	if !s.IsOwner(ctx.EffectiveMessage) {
		return ext.ContinueGroups
	}

	if err := s.SendLogsToOwner(); err != nil {
		s.Logger.Error("can't send logs", zap.Error(err))
	}
	return ext.EndGroups
}

func Undefined(s *mux.Sender, ctx *ext.Context) error {
	// dont send error outside of private chat
	if ctx.EffectiveChat.Type != "private" {
		return ext.EndGroups
	}

	// let's hope ext.Context will be updated some day..
	cntxt := context.Background()

	_, err := s.ReplyToMessage(
		cntxt,
		ctx.EffectiveMessage,
		s.Resp.Get(tr.CmdUndefined, tr.GetLang(ctx.Message)),
		nil,
	)
	if err != nil {
		s.Logger.Error("failed to send Undefined command", zap.Error(err))
	}

	return ext.EndGroups
}

func ProcessURL(s *mux.Sender, ctx *ext.Context) error {
	scloud.ProcessURL(s, ctx)
	return ext.EndGroups
}

func Default(s *mux.Sender, ctx *ext.Context) error {
	// dont send error outside of private chat
	if ctx.EffectiveChat.Type != "private" {
		return ext.EndGroups
	}

	// let's hope ext.Context will be updated some day..
	cntxt := context.Background()

	_, err := s.ReplyToMessage(
		cntxt,
		ctx.EffectiveMessage,
		s.Resp.Get(tr.ErrNotURL, tr.GetLang(ctx.Message)),
		nil,
	)
	if err != nil {
		s.Logger.Error("failed to send Undefined command", zap.Error(err))
	}

	return ext.EndGroups
}
