package scloud

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"net/url"
	"unicode/utf16"
)

func getValidURL(msg *gotgbot.Message) string {
	urls := getURLs(msg)
	for _, u := range urls {
		_, err := url.Parse(u)
		if err != nil {
			continue
		}
		return u
	}

	return ""
}

func getURLs(msg *gotgbot.Message) []string {
	result := make([]string, 0)

	for _, ent := range msg.Entities {
		if ent.Type != "url" {
			continue
		}
		result = append(result, convertToUTF(msg.Text, ent.Offset, ent.Length))
	}
	return result
}

func convertToUTF(value string, start, length int64) string {
	// weird workaround because telegram count symbols in utf16
	text := utf16.Encode([]rune(value))
	return string(utf16.Decode(text[start:][:length]))
}
