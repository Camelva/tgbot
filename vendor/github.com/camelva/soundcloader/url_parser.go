package soundcloader

import (
	"fmt"
	"net/url"
	"regexp"
)

var (
	// General url schema
	generalRE = regexp.MustCompile(
		`(?P<protocol>(?:[a-z]{3,6}://)|(?:^|\s))(?P<host>(?:[a-zA-Z0-9\-]+\.)+[a-z]{2,13})(?P<path>[.?=&%/\w\-:]*\b)`)

	// Different SoundCloud url types
	stationRE = regexp.MustCompile(
		`^/(?:stations)/(?:track)/(?P<user>[\w-]+)/(?P<title>[\w-]+)(?:|/|/(?P<secret>[\w-]+)/?)$`)
	playlistRE = regexp.MustCompile(
		`^/(?P<user>[\w-]+)/(?:sets)/(?P<title>[\w-:]+)(?:|/|/(?P<secret>[\w-]+)/?)$`)
	userRE = regexp.MustCompile(
		`^/(?P<user>[\w-]+)/?$`)
	songRE = regexp.MustCompile(
		`^/(?P<user>[\w-]+)/(?P<title>[\w-]+)(?:|/|/(?P<secret>[\w-]+)/?)$`)
)

type URLInfo struct {
	User        string
	Title       string
	Kind        string
	secretToken string
}

func (s *URLInfo) String() string {
	var result string
	switch s.Kind {
	case "station":
		result = fmt.Sprintf("https://soundcloud.com/stations/track/%s/%s", s.User, s.Title)
	case "playlist":
		result = fmt.Sprintf("https://soundcloud.com/%s/sets/%s", s.User, s.Title)
	case "user":
		result = fmt.Sprintf("https://soundcloud.com/%s", s.User)
	case "song":
		result = fmt.Sprintf("https://soundcloud.com/%s/%s", s.User, s.Title)
	}

	if result == "" {
		return ""
	}

	if s.secretToken != "" {
		return fmt.Sprintf("%s/%s", result, s.secretToken)
	}

	return result
}

func Parse(message string) *URLInfo {
	return DefaultClient.Parse(message)
}

func (c *Client) Parse(message string) *URLInfo {
	u := extractURL(message)
	if u == nil {
		return nil
	}

	return c.ParseURL(u)
}

func ParseURL(u *url.URL) *URLInfo {
	return DefaultClient.ParseURL(u)
}

func (c *Client) ParseURL(u *url.URL) *URLInfo {
	u = c.unwrapURL(u)

	if u.Host != "soundcloud.com" {
		return nil
	}

	var kind string
	var result, names []string
	urlPath := u.EscapedPath()

	urlKinds := map[string]*regexp.Regexp{
		"station":  stationRE,
		"playlist": playlistRE,
		"user":     userRE,
		"song":     songRE,
	}
	for patternName, pattern := range urlKinds {
		tempResult := pattern.FindStringSubmatch(urlPath)
		if tempResult == nil {
			continue
		}

		kind, result, names = patternName, tempResult, pattern.SubexpNames()
		break
	}

	if result == nil || names == nil {
		return nil
	}

	var info = &URLInfo{Kind: kind}

	for i, n := range names {
		if n == "user" {
			info.User = result[i]
		} else if n == "title" {
			info.Title = result[i]
		} else if n == "secret" {
			info.secretToken = result[i]
		}
	}

	return info
}

func extractURL(message string) *url.URL {
	rawURL := generalRE.FindString(message)
	if rawURL == "" {
		return nil
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil
	}
	return u
}

func (c *Client) unwrapURL(u *url.URL) *url.URL {
	if u.Host == "m.soundcloud.com" {
		u.Host = "soundcloud.com"
		return u
	}

	if u.Host != "soundcloud.app.goo.gl" {
		return u
	}

	resp, err := c.httpGet(u.String(), false)
	if err != nil {
		return u
	}
	resp.Body.Close()
	return resp.Request.URL
}
