package telemetry

import (
	"bytes"
	"encoding/json"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"golang.org/x/xerrors"
	"io/ioutil"
	"net/http"
)

var server string

type Client struct {
	server string
}

func New(server string) *Client {
	return &Client{server: server}
}

func (c *Client) Report(msg *gotgbot.Message, success bool) error {
	if c.server == "" {
		// send nothing
		return nil
	}

	chat := buildChat(msg)
	user := buildUser(msg)

	message := Message{
		User:    user,
		Chat:    chat,
		Text:    msg.Text,
		Date:    msg.Date,
		Success: success,
	}

	r := Report{
		Method: "newMessage",
		Args:   message,
	}

	return c.sendReport(r)
}

func buildChat(msg *gotgbot.Message) Chat {
	return Chat{
		ID:        msg.Chat.Id,
		Type:      msg.Chat.Type,
		Title:     msg.Chat.Title,
		Username:  msg.Chat.Username,
		FirstName: msg.Chat.FirstName,
		LastName:  msg.Chat.LastName,
	}
}

func buildUser(msg *gotgbot.Message) User {
	u := User{}
	if msg.Chat.Type != "channel" && msg.From != nil {
		u.ID = msg.From.Id
		u.Username = msg.From.Username
		u.FirstName = msg.From.FirstName
		u.LastName = msg.From.LastName
		u.Language = msg.From.LastName
	}
	return u
}

func (c *Client) sendReport(r Report) error {
	rEncoded, err := json.Marshal(r)
	if err != nil {
		return xerrors.Errorf("can't marshal report: %w", err)
	}

	var userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:73.0) Gecko/20100101 Firefox/73.0"
	req, err := http.NewRequest(http.MethodPost, server, bytes.NewReader(rEncoded))
	if err != nil {
		return xerrors.Errorf("can't build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return xerrors.Errorf("can't send report: %w", err)
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("can't read response: %w", err)
	}
	defer resp.Body.Close()

	serverResp := new(ServerResponse)
	if err := json.Unmarshal(respBody, serverResp); err != nil {
		return err
	}

	if !serverResp.Status {
		return xerrors.Errorf("bad request: %w", err)
	}
	return nil
}

type Report struct {
	Method string  `json:"m"`
	Args   Message `json:"args"`
}

type Message struct {
	User    User   `json:"user"`
	Chat    Chat   `json:"chat"`
	Text    string `json:"text"`
	Date    int64  `json:"date"`
	Success bool   `json:"success"`
}

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Language  string `json:"language"`
}

type Chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type ServerResponse struct {
	Status bool   `json:"status"`
	Data   string `json:"message"`
}
