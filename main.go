package main

import (
	"github.com/camelva/erzo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"strconv"
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
		responseMsg = BotPhrase.ErrTooLarge()
		errCode = 19
	}

	if err != nil {
		reportError(dbMsgID, err, errCode)
	}

	if msg.Chat.Type != "private" {
		return
	}

	sendMessage(bot, msg, responseMsg, true)
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
		sendMessage(bot, message, response, true)
		return nil
	}

	songInfo, err := erzo.GetInfo(message.Text, erzo.OptionTruncate(true))

	if err != nil {
		// if its not url - just ignore in groups but respond in private
		if _, ok := err.(erzo.ErrNotURL); ok {
			if isPrivateChat {
				sendMessage(bot, message, BotPhrase.ErrNotURL(), false)
			}
			return nil
		}
		return err
	}

	if oversized(songInfo.URL) {
		sendMessage(bot, message, BotPhrase.ErrTooLarge(), true)
		return nil
	}

	tmpMessageID = sendMessage(bot, message, BotPhrase.ProcessStart(), true)
	defer deleteMessage(bot, message.Chat, tmpMessageID)

	songData, err := songInfo.Get()
	if err != nil {
		return err
	}
	log.Println("Downloaded song. Uploading to user...")

	tmpMessageID = editMessage(bot, message, BotPhrase.ProcessUploading())

	// Inform user about uploading
	_, _ = bot.Send(tgbotapi.NewChatAction(chatID, "upload_audio"))

	audioMsg := tgbotapi.NewAudioUpload(chatID, songData.Path)
	audioMsg.Title = songData.Title
	audioMsg.Performer = songData.Author
	audioMsg.Duration = int(songData.Duration.Seconds())
	audioMsg.ReplyToMessageID = message.MessageID

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

func sendMessage(bot *tgbotapi.BotAPI, msgInfo *tgbotapi.Message, text string, reply bool) (messageID int) {
	msgObj := tgbotapi.NewMessage(msgInfo.Chat.ID, text)
	if reply {
		msgObj.ReplyToMessageID = msgInfo.MessageID
	}
	return send(bot, msgObj)
}

func editMessage(bot *tgbotapi.BotAPI, msgInfo *tgbotapi.Message, text string) (messageId int) {
	msgObj := tgbotapi.NewEditMessageText(msgInfo.Chat.ID, msgInfo.MessageID, text)
	return send(bot, msgObj)
}

func send(bot *tgbotapi.BotAPI, c tgbotapi.Chattable) (messageID int) {
	for range make([]int, 2) {
		sentMsg, err := bot.Send(c)
		if err != nil {
			continue
		}
		return sentMsg.MessageID
	}
	return 0
}

func deleteMessage(bot *tgbotapi.BotAPI, chat *tgbotapi.Chat, messageID int) {
	msgToDelete := tgbotapi.NewDeleteMessage(chat.ID, messageID)
	if _, err := bot.DeleteMessage(msgToDelete); err != nil {
		log.Printf("error while deleting temp message: %s", err)
	}
}

func oversized(s string) bool {
	r, _ := http.NewRequest(http.MethodGet, s, nil)
	r.Header.Set("Accept-Encoding", "gzip, deflate, br")
	r.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	r.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:80.0) Gecko/20100101 Firefox/80.0")
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return true
	}
	_ = resp.Body.Close()
	sizeBytes, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return true
	}
	sizeMB := sizeBytes / 1024 / 1024
	if sizeMB > 49 {
		return true
	}
	return false
}
