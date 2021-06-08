package mux

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/cenkalti/backoff/v4"
	"os"
	"strconv"
)

func IsTelegramError(err error) bool {
	_, ok := err.(*gotgbot.TelegramError)
	return ok
}

func CheckError(err error) error {
	if err == nil {
		return nil
	}

	if IsTelegramError(err) {
		return err
	} else {
		return backoff.Permanent(err)
	}
}

func GetUserID(msg *gotgbot.Message) int64 {
	if msg.From == nil {
		return msg.Chat.Id
	}
	return msg.From.Id
}

func GetOwner() int64 {
	owner, err := strconv.ParseInt(os.Getenv("BOT_OWNER"), 10, 64)
	if err != nil {
		return 0
	}
	return owner
}

func (s *Sender) IsOwner(msg *gotgbot.Message) bool {
	//owner := GetOwner()
	if s.OwnerID == 0 {
		return false
	}

	return GetUserID(msg) == s.OwnerID
}
