package mux

import (
	"context"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"
	"os"
)

func (s *Sender) SendMessage(
	ctx context.Context,
	msg *gotgbot.Message,
	text string,
	opts *gotgbot.SendMessageOpts,
) (m *gotgbot.Message, err error) {
	if opts == nil {
		opts = &gotgbot.SendMessageOpts{ParseMode: s.ParseMode}
	}

	err = backoff.RetryNotify(
		func() error {
			m, err = s.Client.SendMessage(msg.Chat.Id, text, opts)
			return CheckError(err)
		},
		backoff.WithMaxRetries(backoff.WithContext(backoff.NewExponentialBackOff(), ctx), s.MaxRetries),
		s.LogRetry,
	)

	if err != nil {
		s.Logger.Error("sendMessage failed after retries", zap.Uint64("retry", s.MaxRetries), zap.Error(err))
	}

	return
}

func (s *Sender) ReplyToMessage(
	ctx context.Context,
	msg *gotgbot.Message,
	text string,
	opts *gotgbot.SendMessageOpts,
) (m *gotgbot.Message, err error) {
	if opts == nil {
		opts = &gotgbot.SendMessageOpts{ParseMode: s.ParseMode}
	}

	opts.ReplyToMessageId = msg.MessageId

	return s.SendMessage(ctx, msg, text, opts)
}

func (s *Sender) EditMessage(
	ctx context.Context,
	msg *gotgbot.Message,
	text string,
	opts *gotgbot.EditMessageTextOpts,
) (m *gotgbot.Message, err error) {
	if opts == nil {
		opts = &gotgbot.EditMessageTextOpts{ParseMode: s.ParseMode}
	}

	opts.ChatId = msg.Chat.Id
	opts.MessageId = msg.MessageId

	err = backoff.RetryNotify(
		func() error {
			m, err = s.Client.EditMessageText(text, opts)
			return CheckError(err)
		},
		backoff.WithMaxRetries(backoff.WithContext(backoff.NewExponentialBackOff(), ctx), s.MaxRetries),
		s.LogRetry,
	)

	if err != nil {
		s.Logger.Error("editMessage failed after retries", zap.Uint64("retry", s.MaxRetries), zap.Error(err))
	}

	return
}

func (s *Sender) DeleteMessage(
	ctx context.Context,
	msg *gotgbot.Message,
) (ok bool, err error) {
	err = backoff.RetryNotify(
		func() error {
			ok, err = s.Client.DeleteMessage(msg.Chat.Id, msg.MessageId)
			return CheckError(err)
		},
		backoff.WithMaxRetries(backoff.WithContext(backoff.NewExponentialBackOff(), ctx), s.MaxRetries),
		s.LogRetry,
	)

	if err != nil {
		s.Logger.Error("deleteMessage failed after retries", zap.Uint64("retry", s.MaxRetries), zap.Error(err))
	}
	return
}

func (s *Sender) SendAudio(
	ctx context.Context,
	chat int64,
	f *os.File,
	opts *gotgbot.SendAudioOpts,
) (m *gotgbot.Message, err error) {
	if opts == nil {
		opts = &gotgbot.SendAudioOpts{}
	}

	opts.ParseMode = s.ParseMode

	err = backoff.RetryNotify(
		func() error {
			m, err = s.Client.SendAudio(chat, f, opts)
			return CheckError(err)
		},
		backoff.WithMaxRetries(backoff.WithContext(backoff.NewExponentialBackOff(), ctx), s.MaxRetries),
		s.LogRetry,
	)

	if err != nil {
		s.Logger.Error("sendAudio failed after retries", zap.Uint64("retry", s.MaxRetries), zap.Error(err))
	}

	return
}

func (s *Sender) SendDocument(
	ctx context.Context,
	chat int64,
	f *os.File,
	opts *gotgbot.SendDocumentOpts,
) (m *gotgbot.Message, err error) {
	if opts == nil {
		opts = &gotgbot.SendDocumentOpts{}
	}

	err = backoff.RetryNotify(
		func() error {
			m, err = s.Client.SendDocument(chat, f, opts)
			return CheckError(err)
		},
		backoff.WithMaxRetries(backoff.WithContext(backoff.NewExponentialBackOff(), ctx), s.MaxRetries),
		s.LogRetry,
	)

	if err != nil {
		s.Logger.Error("sendDocument failed after retries", zap.Uint64("retry", s.MaxRetries), zap.Error(err))
	}

	return
}
