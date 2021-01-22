package main

import (
	"bytes"
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io/ioutil"
	"log"
	"net/http"
)

const tServer = "https://bigbonus.pp.ua/api/v2/"

type ServerResponse struct {
	Status        int         `json:"status"`
	StatusMessage string      `json:"status_message"`
	Data          interface{} `json:"data"`
}

type tReport struct {
	Method string `json:"m"`
	Args   tArgs  `json:"args"`
}

type tArgs interface {
	method() string
}

type tMessage struct {
	User tUser  `json:"user"`
	Chat tChat  `json:"chat"`
	Text string `json:"text"`
	Date int    `json:"date"`
}
type tUser struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Language  string `json:"language"`
}
type tChat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (tMessage) method() string { return "newMessage" }

//type tError struct {
//	Text      string `json:"text"`
//	Code      int    `json:"code"`
//	MessageID int    `json:"message_id"`
//}
//
//func (tError) method() string { return "newError" }

func reportMessage(msg *tgbotapi.Message) int {
	var usr = tUser{}
	var cht = tChat{}
	if msg.From != nil {
		usr = tUser{
			ID:        msg.From.ID,
			Username:  msg.From.UserName,
			FirstName: msg.From.FirstName,
			LastName:  msg.From.LastName,
			Language:  msg.From.LanguageCode,
		}
	}
	if msg.Chat != nil {
		cht = tChat{
			ID:        msg.Chat.ID,
			Type:      msg.Chat.Type,
			Title:     msg.Chat.Title,
			Username:  msg.Chat.UserName,
			FirstName: msg.Chat.FirstName,
			LastName:  msg.Chat.LastName,
		}
	}
	message := tMessage{
		User: usr,
		Chat: cht,
		Text: msg.Text,
		Date: msg.Date,
	}
	method := message.method()
	report := tReport{
		Method: method,
		Args:   message,
	}
	if res := sendReport(report); res != nil {
		// server return id of record 	or
		// true if record already exist or
		// false if problem occurred
		if resp, ok := res.(map[string]interface{}); ok {
			if msgID := int(resp["message"].(float64)); msgID > 1 {
				return msgID
			}
		}
	}
	return 0
}

func sendReport(report tReport) (responseData interface{}) {
	data, err := json.Marshal(report)
	if err != nil {
		log.Printf("[telemetry.SendReport]Marshal report: %s\n", err)
		return
	}

	var userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:73.0) Gecko/20100101 Firefox/73.0"
	req, err := http.NewRequest("POST", tServer, bytes.NewReader(data))
	if err != nil {
		log.Printf("[telemetry.SendReport]New request: %s\n", err)
		return
	}
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[telemetry.SendReport]Post request: %s\n", err)
		return
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[telemetry.sendReport]Read response: %s\n", err)
		return
	}
	defer resp.Body.Close()
	serverResponse := new(ServerResponse)

	if err := json.Unmarshal(respBody, serverResponse); err != nil {
		log.Printf("[telemetry.SendReport]Unmarhsal response: %s\n", err)
	}

	if serverResponse.Status == http.StatusBadRequest {
		log.Printf("[telemetry.SendReport]Server response: %s\n", serverResponse.StatusMessage)
		return
	}

	if serverResponse.Status == http.StatusOK {
		return serverResponse.Data
	}

	// unexpected response status
	log.Printf(
		"[telemetry.SendReport]Unexpeted response. Code: %d, message: %s\n",
		serverResponse.Status, serverResponse.StatusMessage)
	return
}
