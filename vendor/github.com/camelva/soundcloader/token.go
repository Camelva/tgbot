package soundcloader

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
)

func (c *Client) getToken() error {
	if c.tokenLocationFile != "" {
		t, err := readTokenFromFile(c.tokenLocationFile)
		if err == nil {
			c.token = t
			return nil
		}
	}

	return c.updateToken()
}

func (c *Client) updateToken() error {
	if c.tokenLocationFile == "" {
		c.tokenLocationFile = filepath.Join(os.TempDir(), "soundcloud-token.txt")
	}

	var scriptRE = regexp.MustCompile(`<script[^>]+src="([^"]+)"`)
	var clientRE = regexp.MustCompile(`client_id\s*:\s*"([0-9a-zA-Z]{32})"`)

	res, err := c.fetch("https://soundcloud.com", false)
	if err != nil {
		return fmt.Errorf("can't fetch soundcloud.com: %s", err)
	}

	scripts := scriptRE.FindAllStringSubmatch(string(res), -1)

	for _, script := range scripts {
		scriptURL, err := url.Parse(script[1])
		if err != nil {
			// can't parse script url. Let's try next one
			continue
		}
		scriptBody, err := c.fetch(scriptURL.String(), false)
		if err != nil {
			// can't fetch script. Let's try next one
			continue
		}
		matches := clientRE.FindSubmatch(scriptBody)
		if matches == nil {
			continue
		}

		// save token
		c.token = string(matches[1])
		_ = ioutil.WriteFile(c.tokenLocationFile, matches[1], 0755)
		return nil
	}
	return fmt.Errorf("can't retrieve token")
}

func readTokenFromFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
