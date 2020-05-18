package main

import (
	"log"

	"github.com/camelva/erzo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var BotPhrase = Responses.Get("en")

func main() {
	config := loadConfig("config.yml")

	bot, err := tgbotapi.NewBotAPI(config.Telegram.Token)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s\n", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		var msg *tgbotapi.Message
		if update.Message != nil {
			msg = update.Message
		} else if update.ChannelPost != nil {
			msg = update.ChannelPost
		} else {
			continue
		}
		dbMsgID := reportMessage(msg)
		if err := handleMessage(bot, msg); err != nil {
			handleError(bot, msg, err, dbMsgID)
		}
	}
}

func handleError(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, err error, dbMsgID int) {
	var responseMsg string
	var errCode int
	switch err.(type) {
	case erzo.ErrNotURL:
		responseMsg = BotPhrase.ErrNotURL()
		err = nil // its not error
		log.Println("Received message without link. Responding...")
	case erzo.ErrUnsupportedService:
		responseMsg = BotPhrase.ErrUnsupportedService()
		errCode = 10
		log.Println("Received message with link from unsupported service. Responding...")
	case erzo.ErrUnsupportedProtocol:
		// almost similar for user but we need to report about it
		responseMsg = BotPhrase.ErrUnsupportedService()
		errCode = 31
	case erzo.ErrUnsupportedType:
		if err.(erzo.ErrUnsupportedType).Format == "playlist" {
			responseMsg = BotPhrase.ErrPlaylist()
		} else {
			responseMsg = BotPhrase.ErrUnsupportedFormat()
		}
		errCode = 11
		log.Println("Received message with unsupported url type. Responding...")
	case erzo.ErrCantFetchInfo:
		// most of the time, can't fetch if song is unavailable, and that's what we respond to user
		// we don't really need to report this error to analytic, but lets keep it for more verbose
		responseMsg = BotPhrase.ErrUnavailableSong()
		errCode = 20
	case erzo.ErrDownloadingError:
		// it means we fetched all info, but could not download it. Tell user to try again
		responseMsg = BotPhrase.ErrUndefined()
		errCode = 30
	case erzo.ErrUndefined:
		responseMsg = BotPhrase.ErrUndefined()
		errCode = 99
	default:
		responseMsg = BotPhrase.ErrUndefined()
		errCode = 99
	}
	if err != nil && err.Error() == "Request Entity Too Large" {
		responseMsg = "Looks like this song weighs too much.\n" +
			"Telegram limits uploading files size to 50mb and we can't avoid this limit.\n" +
			"Please try another one"
		errCode = 19
	}

	if err != nil {
		reportError(dbMsgID, err, errCode)
	}

	if msg.Chat.Type != "private" {
		return
	}

	sendMessage(bot, msg.Chat.ID, responseMsg)
}

func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	// Update responses language first
	language := "en"
	if message.From != nil {
		language = message.From.LanguageCode
	}
	BotPhrase = Responses.Get(language)

	var isPrivateChat = message.Chat.Type == "private"
	var tmpMessageID int
	var chatID = message.Chat.ID

	log.Println("Received new message", message.Text)

	if message.IsCommand() {
		response := checkForCommands(message)
		sendMessage(bot, chatID, response)
		return nil
	}

	if isPrivateChat {
		tmpMessageID = sendMessage(bot, chatID, BotPhrase.ProcessStart())
	}
	defer deleteTempMessage(bot, message.Chat, tmpMessageID)

	songFile, err := erzo.Get(message.Text, erzo.Truncate(true))
	if err != nil {
		return err
	}
	log.Println("Downloaded song. Uploading to user...")

	if isPrivateChat {
		// update old temp message
		tmpMessageID = sendMessage(bot, chatID, BotPhrase.ProcessUploading(), tmpMessageID)
	}

	// Inform user about uploading
	// but we don't care about possible errors
	_, _ = bot.Send(tgbotapi.NewChatAction(chatID, "upload_audio"))

	audioMsg := tgbotapi.NewAudioUpload(chatID, songFile)

	// and only then send song file
	if _, err := bot.Send(audioMsg); err != nil {
		return err
	}

	log.Println("Waiting for another message ~_~")
	return nil
}

func checkForCommands(message *tgbotapi.Message) (response string) {
	log.Println("Received command -", message.Text, ".Responding...")

	switch message.Command() {
	case "help":
		return BotPhrase.CmdHelp()
	case "start":
		return BotPhrase.CmdStart()
	default:
		return BotPhrase.CmdUnknown()
	}
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string, oldMsgContainer ...int) (msgID int) {
	var msgObj tgbotapi.Chattable
	if oldMsgContainer != nil {
		// Edit old message instead of creating new
		oldMsgID := oldMsgContainer[0]
		msgObj = tgbotapi.NewEditMessageText(chatID, oldMsgID, text)
	} else {
		msgObj = tgbotapi.NewMessage(chatID, text)
	}
	// two tries
	for range make([]int, 2) {
		sentMsg, err := bot.Send(msgObj)
		if err != nil {
			continue
		}
		return sentMsg.MessageID
	}
	return 0
}

func deleteTempMessage(bot *tgbotapi.BotAPI, chat *tgbotapi.Chat, messageID int) {
	if chat.Type != "private" {
		return
	}
	msgToDelete := tgbotapi.NewDeleteMessage(chat.ID, messageID)
	if _, err := bot.DeleteMessage(msgToDelete); err != nil {
		log.Printf("error while deleting temp message: %s", err)
	}
}
