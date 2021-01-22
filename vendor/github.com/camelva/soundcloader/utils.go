package soundcloader

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func oneOf(s1, s2 string) string {
	if s1 != "" {
		return s1
	}
	return s2
}

func alreadyExist(fileLocation string) bool {
	_, err := os.Stat(fileLocation)
	if err != nil {
		return false
	}
	return true
}

func logRequest(req *http.Request, body bool) {
	out, err := httputil.DumpRequest(req, body)
	if err != nil {
		log.Print(err)
		return
	}
	log.Print(string(out))
}

//func logResponse(resp *http.Response, body bool) {
//	out, err := httputil.DumpResponse(resp, body)
//	if err != nil {
//		log.Print(err)
//		return
//	}
//	log.Print(string(out))
//}

func filenameFromHeader(header http.Header) string {
	var title string
	tmp := header.Get("Content-Disposition")
	ss := strings.Split(tmp, ";")
	for _, s := range ss {
		s := strings.TrimSpace(s)
		if strings.HasPrefix(s, "filename=") {
			title = strings.TrimPrefix(s, "filename=")
			break
		} else if strings.HasPrefix(s, `filename*=utf-8''`) {
			title = strings.TrimPrefix(s, `filename*=utf-8''`)
			break
		}
	}
	title = strings.Trim(title, `"`)
	unescapedTitle, err := url.QueryUnescape(title)
	if err != nil {
		return title
	}
	return unescapedTitle
}

func (c *Client) downloadThumbnail(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	name := u.EscapedPath()
	fileLocation := filepath.Join(c.OutputFolder, name)

	content, err := c.fetch(s, false)
	if err != nil {
		return "", err
	}

	if err := ioutil.WriteFile(fileLocation, content, 0755); err != nil {
		return "", err
	}
	return fileLocation, nil
}
