package main

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters"
	"log"
	"net/http"
	"os"
	"tgbot/telemetry"
	"time"
)

func main() {
	config := loadConfigs("env.yml", "config.yml")

	b, err := gotgbot.NewBot(config.Telegram.Token, &gotgbot.BotOpts{
		Client:      http.Client{},
		GetTimeout:  time.Second * 10,
		PostTimeout: time.Minute * 5,
	})
	if err != nil {
		panic("failed to create new bot: " + err.Error())
	}

	// need to init telemetry
	telemetry.SetServer("https://bigbonus.pp.ua/api/v2/")

	// Create updater and dispatcher.
	updater := ext.NewUpdater(&ext.UpdaterOpts{
		ErrorLog: log.New(os.Stderr, "ERROR", log.LstdFlags),
		DispatcherOpts: ext.DispatcherOpts{
			Error:    onError,
			ErrorLog: log.New(os.Stderr, "ERROR", log.LstdFlags),
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
		panic("failed to start polling: " + err.Error())
	}
	fmt.Printf("%s has been started...\n", b.User.Username)

	// Idle, to keep updates coming in, and avoid bot stopping.
	updater.Idle()
}

func onError(ctx *ext.Context, err error) {
	log.Printf("%s while doing %s", err, ctx.EffectiveMessage.Text)
}
