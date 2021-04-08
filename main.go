package main

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters"
	"github.com/sirupsen/logrus"
	stdLog "log"
	"net/http"
	"os"
	"tgbot/telemetry"
	"time"
)

var log *logrus.Logger

func main() {
	log = initLog()

	config := loadConfigs("env.yml", "config.yml")

	b, err := gotgbot.NewBot(config.Telegram.Token, &gotgbot.BotOpts{
		Client:      http.Client{},
		GetTimeout:  time.Second * 10,
		PostTimeout: time.Minute * 5,
	})
	if err != nil {
		log.WithError(err).Fatal("failed to create new bot")
	}

	// need to init telemetry
	telemetry.SetServer("https://bigbonus.pp.ua/api/v2/")

	// Create updater and dispatcher.
	updater := ext.NewUpdater(&ext.UpdaterOpts{
		ErrorLog: stdLog.New(os.Stderr, "ERROR", stdLog.LstdFlags),
		DispatcherOpts: ext.DispatcherOpts{
			Error:    onError,
			ErrorLog: stdLog.New(os.Stderr, "ERROR", stdLog.LstdFlags),
		}})
	dispatcher := updater.Dispatcher

	// Commands first
	dispatcher.AddHandlerToGroup(handlers.NewCommand("start", cmdStart), 0)
	dispatcher.AddHandlerToGroup(handlers.NewCommand("help", cmdHelp), 0)
	dispatcher.AddHandlerToGroup(handlers.NewMessage(filters.Command, cmdUndefined), 0)

	// Then messages with url
	dispatcher.AddHandlerToGroup(handlers.NewMessage(filters.Entity("url"), checkURL), 1)

	// Last - everything else
	dispatcher.AddHandlerToGroup(handlers.NewMessage(filters.All, replyNotURL), 1)

	// Start receiving updates.
	err = updater.StartPolling(b, &ext.PollingOpts{})
	if err != nil {
		log.WithError(err).Fatal("failed to start polling: ")
	}
	log.Info("%s has been started...\n", b.User.Username)

	// Idle, to keep updates coming in, and avoid bot stopping.
	updater.Idle()
}

func onError(ctx *ext.Context, err error) {
	log.WithError(err).Error(ctx.EffectiveMessage.Text)
}
