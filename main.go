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
	"os/signal"
	"tgbot/storage"
	"tgbot/telemetry"
	"time"
)

var log *logrus.Logger

var OwnerID int64

func main() {
	log = initLog()

	config := loadConfigs(".env")
	storage.SetToken(config.Storage.Token)
	OwnerID = config.Telegram.OwnerID

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

	setHandlers(dispatcher)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		_ = sendLogs(b)
		_ = updater.Stop()
	}()

	// Start receiving updates.
	err = updater.StartPolling(b, &ext.PollingOpts{})
	if err != nil {
		log.WithError(err).Fatal("failed to start polling: ")
	}
	log.Infof("%s has been started...", b.User.Username)

	// Idle, to keep updates coming in, and avoid bot stopping.
	updater.Idle()
}

func onError(ctx *ext.Context, err error) {
	log.WithError(err).Error(ctx.EffectiveMessage.Text)
}

func setHandlers(dispatcher *ext.Dispatcher) {
	// Global logging
	hLogAll := handlers.Message{AllowChannel: true, Filter: filters.All, Response: logMessage}
	dispatcher.AddHandlerToGroup(hLogAll, 0)

	// Commands
	// - first log every command received
	hLogCmd := handlers.Message{AllowChannel: true, Filter: filters.Command, Response: logCmd}
	dispatcher.AddHandlerToGroup(hLogCmd, 1)

	hStart := handlers.Command{Triggers: []rune{'/'}, AllowChannel: true, Command: "start", Response: cmdStart}
	dispatcher.AddHandlerToGroup(hStart, 1)

	hHelp := handlers.Command{Triggers: []rune{'/'}, AllowChannel: true, Command: "help", Response: cmdHelp}
	dispatcher.AddHandlerToGroup(hHelp, 1)

	hLogs := handlers.Command{Triggers: []rune{'/'}, AllowChannel: true, Command: "logs", Response: cmdLogs}
	dispatcher.AddHandlerToGroup(hLogs, 1)

	// - if no match within defined commands, but ignore if not private chat
	hUndefined := handlers.Message{AllowChannel: false, Filter: filters.Command, Response: cmdUndefined}
	dispatcher.AddHandlerToGroup(hUndefined, 1)

	// Messages containing url
	hURL := handlers.Message{AllowChannel: true, Filter: filters.Entity("url"), Response: checkURL}
	dispatcher.AddHandlerToGroup(hURL, 2)

	// Lastly - default reply if nothing match. Ignore channels (and groups too, but inside of handler)
	hDefault := handlers.Message{AllowChannel: false, Filter: filters.All, Response: replyNotURL}
	dispatcher.AddHandlerToGroup(hDefault, 2)
}
