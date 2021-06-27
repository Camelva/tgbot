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

	AppCtx context.Context

	LogRetry func(error, time.Duration)

	MaxRetries uint64
	ParseMode  string

	Resp     *tr.Translation
	External *storage.Storage

	Stats   *telemetry.Client
	Logger  *zap.Logger
	LogFile string
}

type Options struct {
	Bot     *gotgbot.Bot
	OwnerID int64

	AppContext context.Context

	LogRetry func(error, time.Duration)

	MaxRetries uint64
	ParseMode  string

	Translation     *tr.Translation
	ExternalStorage *storage.Storage

	Telemetry *telemetry.Client
	Logger    *zap.Logger
	LogFile   string
}

func New(opts Options) *Sender {
	if opts.Bot == nil || opts.Translation == nil ||
		opts.ExternalStorage == nil || opts.Telemetry == nil {
		return nil
	}

	if opts.MaxRetries == 0 {
		opts.MaxRetries = 7
	}

	if opts.ParseMode == "" {
		opts.ParseMode = "HTML"
	}

	if opts.Logger == nil {
		opts.Logger = zap.NewNop()
	}

	if opts.OwnerID == 0 {
		opts.OwnerID = GetOwner()
	}

	s := &Sender{
		Client:     opts.Bot,
		AppCtx:     opts.AppContext,
		ParseMode:  opts.ParseMode,
		MaxRetries: opts.MaxRetries,
		External:   opts.ExternalStorage,
		Resp:       opts.Translation,
		Stats:      opts.Telemetry,
		Logger:     opts.Logger,
		LogFile:    opts.LogFile,
	}

	if opts.LogRetry == nil {
		opts.LogRetry = DefaultLogRetry(s)
	}

	s.LogRetry = opts.LogRetry
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
