package soundcloader

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

var (
	EmptyStream      = errors.New("empty stream")
	NoOriginalStream = errors.New("original stream are not provided")

	NotURL  = errors.New("not url")
	NotSong = errors.New("not a song")

	InvalidURL  = errors.New("invalid url")
	InvalidJSON = errors.New("can't unmarshal json")
)

type Client struct {
	HTTPClient   http.Client
	OutputFolder string
	Debug        bool

	token             string
	tokenLocationFile string
}

var DefaultClient = &Client{
	HTTPClient:        http.Client{Timeout: 3 * time.Minute},
	tokenLocationFile: filepath.Join(os.TempDir(), "soundcloud-token.txt"),
	OutputFolder:      filepath.Join(".", "out"),
}

func (c *Client) SetDebug(b bool) {
	c.Debug = b
}

func Get(s string) (*Song, error) {
	return DefaultClient.Get(s)
}

func (c *Client) Get(s string) (*Song, error) {
	info := Parse(s)
	song, err := c.GetURL(info)
	return song, err
}

func GetURL(u *URLInfo) (*Song, error) {
	return DefaultClient.GetURL(u)
}

func (c *Client) GetURL(u *URLInfo) (song *Song, err error) {
	if u == nil {
		return nil, NotURL
	}

	// temporary guard
	if u.Kind != "song" {
		return nil, NotSong
	}

	if c.token == "" {
		if err = c.getToken(); err != nil {
			return
		}
	}

	meta, err := c.getMetadata(u)
	if err != nil {
		return
	}

	song = &Song{client: c}
	song.parseSongInfo(meta)
	return
}

func (c *Client) getMetadata(s *URLInfo) (*metadataV2, error) {
	cleanURL := s.String()

	resolveURL := fmt.Sprintf("https://api-v2.soundcloud.com/resolve?url=%s", cleanURL)

	respData, err := c.fetch(resolveURL, true)
	if err != nil {
		return nil, err
	}

	// if response = "{}"
	if len(respData) < 3 {
		return nil, fmt.Errorf("no metadata, URL: %s", cleanURL)
	}

	meta := new(metadataV2)
	if err := json.Unmarshal(respData, meta); err != nil {
		return nil, InvalidJSON
	}

	// update DownloadURL field
	meta.DownloadURL = fmt.Sprintf("https://api-v2.soundcloud.com/tracks/%d/download", meta.ID)
	return meta, nil
}

func (c *Client) fetch(u string, withToken bool) ([]byte, error) {
	resp, err := c.httpGet(u, withToken)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func (c *Client) httpGet(u string, withToken bool) (*http.Response, error) {
	if withToken {
		if err := c.addClientID(&u); err != nil {
			return nil, err
		}
	}

	var userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:73.0) Gecko/20100101 Firefox/73.0"

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)

	if c.Debug {
		logRequest(req, false)
	}
	return c.HTTPClient.Do(req)
}

func (c *Client) addClientID(uri *string) error {
	u, err := url.Parse(*uri)
	if err != nil {
		return InvalidURL
	}

	q := u.Query()

	if c.token == "" {
		err := c.getToken()
		if err != nil {
			return err
		}
	}

	q.Set("client_id", c.token)
	u.RawQuery = q.Encode()
	*uri = u.String()
	return nil
}

func (c *Client) nativeLoad(fileLocation string, uri string, useOriginalName bool) (string, error) {
	resp, err := c.httpGet(uri, false)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if useOriginalName {
		name := filenameFromHeader(resp.Header)
		fileLocation = filepath.Join(fileLocation, name)

		if alreadyExist(fileLocation) {
			return fileLocation, nil
		}
	}

	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(fileLocation, res, 0755)
	return fileLocation, err
}

func (c *Client) ffmpegLoad(fileLocation string, uri string) (string, error) {
	if err := c.ffmpegGet(uri, fileLocation); err != nil {
		return "", err
	}
	return fileLocation, nil
}
