package mux

import (
	"context"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
	"os"
	tr "tgbot/internal/resp"
	"tgbot/internal/storage"
	"tgbot/internal/telemetry"
	"time"
)

type Sender struct {
	Client  *gotgbot.Bot
	OwnerID int64

	LogRetry func(error, time.Duration)

	MaxRetries uint64
	ParseMode  string

	Resp     *tr.Translation
	External *storage.Storage

	Stats   *telemetry.Client
	Logger  *zap.Logger
	LogFile string
}

func New(
	b *gotgbot.Bot,
	t *telemetry.Client,
	eStorage *storage.Storage,
	resp *tr.Translation,
	l *zap.Logger,
	logFile string,
) *Sender {
	s := &Sender{
		Client:     b,
		ParseMode:  "HTML",
		MaxRetries: 7,
		External:   eStorage,
		Resp:       resp,
		Stats:      t,
		Logger:     l,
		LogFile:    logFile,
	}
	s.LogRetry = DefaultLogRetry(s)
	return s
}

func DefaultLogRetry(s *Sender) func(error, time.Duration) {
	return func(err error, d time.Duration) {
		s.Logger.Debug("retrying request..", zap.Error(err), zap.Duration("wait", d))
	}
}

func (s *Sender) ReportLogs() (rerr error) {
	if s.OwnerID == 0 {
		id := GetOwner()
		if id == 0 {
			return xerrors.New("BOT_OWNER unset")
		}
		s.OwnerID = id
	}

	f, err := os.Open(s.LogFile)
	if err != nil {
		return xerrors.Errorf("can't open log file: %w", err)
	}
	_, err = s.SendDocument(context.Background(), s.OwnerID, f, nil)
	return err
}
