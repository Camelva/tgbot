package main

import (
	"context"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"tgbot/internal/mux"
	tr "tgbot/internal/resp"
	"tgbot/internal/storage"
	"tgbot/internal/telemetry"
	"time"
)

type App struct {
	client *gotgbot.Bot

	updater         *ext.Updater
	mux             *mux.Sender
	externalStorage *storage.Storage
	resp            *tr.Translation

	stats   *telemetry.Client
	logger  *zap.Logger
	logFile string
}

func InitApp(logger *zap.Logger, logFile string) (*App, error) {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		return nil, xerrors.New("no BOT_TOKEN provided")
	}

	externalToken := os.Getenv("STORAGE_TOKEN")
	if token == "" {
		return nil, xerrors.New("no STORAGE_TOKEN provided")
	}

	telemetryServer := os.Getenv("TELEMETRY_SERVER")
	t := telemetry.New(telemetryServer)

	b, err := gotgbot.NewBot(token, &gotgbot.BotOpts{
		Client:      http.Client{},
		GetTimeout:  time.Second * 10,
		PostTimeout: time.Minute * 5, // uploading songs can take some time
	})
	if err != nil {
		return nil, xerrors.Errorf("can't create bot: %w", err)
	}

	updater := ext.NewUpdater(&ext.UpdaterOpts{
		ErrorLog: zap.NewStdLog(logger),
		DispatcherOpts: ext.DispatcherOpts{
			ErrorLog: zap.NewStdLog(logger),
		}})

	external := storage.New(externalToken, logger.Named("External storage"))

	resp := tr.New(filepath.Join("internal", "resp"))

	sender := mux.New(b, t, external, resp, logger, logFile)

	app := &App{
		client:          b,
		updater:         &updater,
		mux:             sender,
		stats:           t,
		externalStorage: external,
		resp:            resp,
		logger:          logger,
		logFile:         logFile,
	}

	return app, nil
}

func (a *App) Close() error {
	a.logger.Sugar().Info("closing..")
	if err := a.mux.SendLogsToOwner(); err != nil {
		a.logger.Error("can't send logs to owner: %w", zap.Error(err))
	}

	// maybe some databases in future..
	return nil
}

func (a *App) Run(ctx context.Context) (err error) {
	if err := startWebhook(a); err != nil {
		if err2 := startPolling(a); err2 != nil {
			return multierr.Append(err, err2)
		}
	}

	a.logger.Info("bot has been started", zap.String("username", a.client.User.Username))

	go a.updater.Idle()

	select {
	case <-ctx.Done():
		a.logger.Info("context was cancelled, stop polling..")
		return a.updater.Stop()
	}
}

func startWebhook(a *App) error {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil || port == 0 {
		port = 443
	}

	domain := os.Getenv("WEBHOOK_DOMAIN")
	if domain == "" {
		return xerrors.New("webhook domain unset")
	}

	opts := ext.WebhookOpts{
		URLPath: a.client.Token,
		Port:    port,
	}
	wHookPath := opts.GetWebhookURL(domain)

	if err := a.updater.StartWebhook(a.client, opts); err != nil {
		return xerrors.Errorf("failed to start webhook: %w", err)
	}

	if _, err := a.client.SetWebhook(wHookPath, nil); err != nil {
		return xerrors.Errorf("failed to set webhook: %w", err)
	}
	return nil
}

func startPolling(a *App) error {
	_, err := a.client.DeleteWebhook(nil)
	if err != nil {
		return xerrors.Errorf("delete webhook failed: %w", err)
	}

	if err := a.updater.StartPolling(a.client, nil); err != nil {
		return xerrors.Errorf("failed to start polling: %w", err)
	}
	return nil
}
