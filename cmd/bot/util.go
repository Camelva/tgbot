package main

import (
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters"
)

func xCommand(c string, r handlers.Response, AllowChannel bool) handlers.Command {
	h := handlers.NewCommand(c, r)
	h.AllowChannel = AllowChannel

	return h
}

func xMessage(f filters.Message, r handlers.Response, AllowChannel bool) handlers.Message {
	h := handlers.NewMessage(f, r)
	h.AllowChannel = AllowChannel

	return h
}
