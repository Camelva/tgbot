package main

import (
	"context"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
	"tgbot/internal/h"
)

func runBot(ctx context.Context, logger *zap.Logger, logFile string) (rerr error) {
	app, err := InitApp(ctx, logger, logFile)
	if err != nil {
		return xerrors.Errorf("initialize: %w", err)
	}
	defer func() {
		multierr.AppendInto(&rerr, app.Close())
	}()

	if err := setupBot(app); err != nil {
		return xerrors.Errorf("setup: %w", err)
	}

	if err := app.Run(ctx); err != nil {
		return xerrors.Errorf("run: %w", err)
	}
	return nil
}

func setupBot(app *App) error {
	// Start with logging
	app.updater.Dispatcher.AddHandlerToGroup(
		xMessage(filters.All,
			func(b *gotgbot.Bot, ctx *ext.Context) error {
				return h.LogMessage(app.mux, ctx)
			}, true),
		0,
	)

	// Commands
	app.updater.Dispatcher.AddHandlerToGroup(
		xMessage(filters.Command,
			func(b *gotgbot.Bot, ctx *ext.Context) error {
				return h.LogCommand(app.mux, ctx)
			}, true),
		1,
	)
	app.updater.Dispatcher.AddHandlerToGroup(
		xCommand("start",
			func(b *gotgbot.Bot, ctx *ext.Context) error {
				return h.Start(app.mux, ctx)
			}, true),
		1,
	)
	app.updater.Dispatcher.AddHandlerToGroup(
		xCommand("help",
			func(b *gotgbot.Bot, ctx *ext.Context) error {
				return h.Help(app.mux, ctx)
			}, true),
		1,
	)
	app.updater.Dispatcher.AddHandlerToGroup(
		xCommand("get",
			func(b *gotgbot.Bot, ctx *ext.Context) error {
				return h.Get(app.mux, ctx)
			}, true),
		1,
	)
	app.updater.Dispatcher.AddHandlerToGroup(
		xCommand("donate",
			func(b *gotgbot.Bot, ctx *ext.Context) error {
				return h.Donate(app.mux, ctx)
			}, true),
		1,
	)

	app.updater.Dispatcher.AddHandlerToGroup(
		xCommand("logs",
			func(b *gotgbot.Bot, ctx *ext.Context) error {
				return h.SendLogs(app.mux, ctx)
			}, false),
		1,
	)

	app.updater.Dispatcher.AddHandlerToGroup(
		xMessage(filters.Command,
			func(b *gotgbot.Bot, ctx *ext.Context) error {
				return h.Undefined(app.mux, ctx)
			}, false),
		1,
	)

	// Messages
	app.updater.Dispatcher.AddHandlerToGroup(
		xMessage(filters.Entity("url"),
			func(b *gotgbot.Bot, ctx *ext.Context) error {
				return h.ProcessURL(app.mux, ctx)
			}, true),
		2,
	)

	app.updater.Dispatcher.AddHandlerToGroup(
		xMessage(filters.All,
			func(b *gotgbot.Bot, ctx *ext.Context) error {
				return h.Default(app.mux, ctx)
			}, false),
		2,
	)

	return nil
}
