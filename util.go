package main

import (
	"bytes"
	"errors"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"io/ioutil"
)

func isOwner(id int64) bool {
	return id == OwnerID
}

func sendLogs(b *gotgbot.Bot) error {
	if OwnerID == 0 {
		return errors.New("can't send logs with zero ownerId")
	}

	data, err := ioutil.ReadFile(LogFile)
	if err != nil {
		return err
	}

	_, err = b.SendDocument(OwnerID, gotgbot.NamedFile{FileName: LogFile, File: bytes.NewReader(data)}, nil)
	return err
}
