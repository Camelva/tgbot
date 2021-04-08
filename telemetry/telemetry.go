package telemetry

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"io/ioutil"
	"net/http"
)

var server string

type Report struct {
	Method string `json:"m"`
	Args Message `json:"args"`
}

type Message struct {
	User User `json:"user"`
	Chat Chat `json:"chat"`
	Text string `json:"text"`
	Date int64 `json:"date"`
	Success bool `json:"success"`
}

type User struct {
	ID int64 `json:"id"`
	Username string `json:"username"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Language string `json:"language"`
}

type Chat struct {
	ID int64 `json:"id"`
	Type string `json:"type"`
	Title string `json:"title"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type ServerResponse struct {
	Status        bool         `json:"status"`
	Data          string `json:"message"`
}

func SetServer(serverURL string) {
	server = serverURL
}

func SendReport(ctx *ext.Context, success bool) error {
	if server == "" {
		return errors.New("you need to set telemetry server first")
	}

	c := Chat{
		ID:        ctx.EffectiveChat.Id,
		Type:      ctx.EffectiveChat.Type,
		Title:     ctx.EffectiveChat.Title,
		Username:  ctx.EffectiveChat.Username,
		FirstName: ctx.EffectiveChat.FirstName,
		LastName:  ctx.EffectiveChat.LastName,
	}

	u := User{}
	if ctx.EffectiveChat.Type != "channel" {
		u.ID = ctx.EffectiveUser.Id
		u.Username = ctx.EffectiveUser.Username
		u.FirstName = ctx.EffectiveUser.FirstName
		u.LastName = ctx.EffectiveUser.LastName
		u.Language = ctx.EffectiveUser.LastName
	}

	m := Message{
		User: u,
		Chat: c,
		Text: ctx.EffectiveMessage.Text,
		Date: ctx.EffectiveMessage.Date,
		Success: success,
	}

	r := Report{
		Method: "newMessage",
		Args:   m,
	}

	rEncoded, err := json.Marshal(r)
	if err != nil {
		return err
	}

	var userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:73.0) Gecko/20100101 Firefox/73.0"
	req, err := http.NewRequest(http.MethodPost, server, bytes.NewReader(rEncoded))
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	serverResp := new(ServerResponse)
	if err := json.Unmarshal(respBody, serverResp); err != nil {
		return err
	}

	if !serverResp.Status {
		return errors.New(serverResp.Data)
	}
	return nil
}
