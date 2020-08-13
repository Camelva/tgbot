package soundcloud

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/camelva/erzo/engine"
	"github.com/camelva/erzo/parsers"
	"github.com/camelva/erzo/utils"
)

// type for identifying url kind
type urlKind int8

// const for identifying user-provided url's type (kind)
const (
	_playlist urlKind = iota + 1
	_station
	_user

	_song // song should be latest element
)

func (k urlKind) String() string {
	switch k {
	case 0:
		return "song"
	case 1:
		return "playlist"
	case 2:
		return "station"
	case 3:
		return "user"
	default:
		return "undefined"
	}
}

var tokenFile = path.Join(os.TempDir(), "soundcloud-token.txt")

var IE extractor

func init() {
	//noinspection SpellCheckingInspection
	clientIDBase := "psT32GLDMZ0TQKgfPkzrGIlco3PYA1kf"

	IE = extractor{
		name:       "SoundCloud",
		urlPattern: `(?:(?:www\.)|(?:m\.)(?:w\.))?soundcloud\.com`,
		apiURL:     "https://api.soundcloud.com/",
		api2URL:    "https://api-v2.soundcloud.com/",
		baseURL:    "https://soundcloud.com/",
		clientID:   clientIDBase,
	}

	if tokenBytes, err := ioutil.ReadFile(tokenFile); err == nil {
		tokenStr := string(tokenBytes)
		if len(tokenStr) == 32 {
			IE.clientID = tokenStr
		}
	}
	engine.AddExtractor(IE)
	return
}

// Main struct with necessary info and methods
type extractor struct {
	name       string
	urlPattern string
	apiURL     string
	api2URL    string
	baseURL    string
	clientID   string
}

func (ie extractor) Name() string {
	return ie.name
}

func (ie extractor) Compatible(u url.URL) bool {
	s := u.Hostname()
	ok, _ := regexp.MatchString(ie.urlPattern, s)
	return ok
}

func (ie extractor) Extract(u url.URL) (*parsers.ExtractorInfo, error) {
	sc := parseURL(u)
	if sc.kind != _song {
		return nil, parsers.ErrFormatNotSupported{Format: sc.kind.String()}
	}
	metadata, err := resolve(sc.url)
	if err != nil || (metadata == &metadata2{}) {
		return nil, parsers.ErrCantContinue{Reason: "can't get song metadata"}
	}
	info, err := extractInfo(metadata)
	if err != nil {
		return nil, parsers.ErrCantContinue{Reason: err.Error()}
	}
	return info, nil
}

// Struct containing info about user-provided url
type scURL struct {
	title  string
	user   string
	kind   urlKind
	secret string
	url    string
}

func parseURL(u url.URL) *scURL {
	urlPath := u.EscapedPath()
	stationTmpl := `^/(?:stations)/(?:track)/([\w-]+)/([\w-]+)(?:|/|/([\w-]+)/?)$`
	stationRE := regexp.MustCompile(stationTmpl)
	playlistTmpl := `^/([\w-]+)/(?:sets)/([\w-]+)(?:|/|/([\w-]+)/?)$`
	playlistRE := regexp.MustCompile(playlistTmpl)
	userTmpl := `^/([\w-]+)/?$`
	userRE := regexp.MustCompile(userTmpl)
	songTmpl := `^/([\w-]+)/([\w-]+)(?:|/|/([\w-]+)/?)$`
	songRE := regexp.MustCompile(songTmpl)
	kinds := map[urlKind]*regexp.Regexp{
		_station:  stationRE,
		_playlist: playlistRE,
		_user:     userRE,
		_song:     songRE,
	}
	for idx, k := range kinds {
		result := k.FindStringSubmatch(urlPath)
		if result == nil {
			continue
		}
		var user, title, secret, uri string
		if len(result) > 1 {
			user = result[1]
		}
		if len(result) > 2 {
			title = result[2]
		}
		if len(result) > 3 {
			secret = result[3]
		}

		switch idx {
		case _station:
			uri = fmt.Sprintf("%sstations/track/%s/%s", IE.baseURL, user, title)
		case _playlist:
			uri = fmt.Sprintf("%ssets/%s/%s", IE.baseURL, user, title)
		case _user:
			uri = fmt.Sprintf("%s%s", IE.baseURL, user)
		case _song:
			// pretty sure its never happened, but still..
			if (user == "stations") || (title == "sets") {
				log.Println("take a look into [soundcloud.go] parseUrl()")
				continue
			}
			uri = fmt.Sprintf("%s%s/%s", IE.baseURL, user, title)
		}
		if secret != "" {
			uri = fmt.Sprintf("%s/%s", uri, secret)
		}
		sc := scURL{
			title:  title,
			user:   user,
			kind:   idx,
			secret: secret,
			url:    uri,
		}
		return &sc
	}
	return &scURL{}
}

func resolve(link string) (*metadata2, error) {
	uri := fmt.Sprintf("%sresolve?url=%s", IE.api2URL, link)
	resolveURL, err := url.Parse(uri)
	if err != nil {
		return nil, parsers.ErrCantContinue{Reason: fmt.Sprintf("can't parse url: %s", uri)}
	}
	res, err := fetch(resolveURL)
	// empty json object "{}"
	if err != nil || (len(res) < 3) {
		return nil, parsers.ErrCantContinue{Reason: "can't fetch resolve url"}
	}
	var scMetadata = new(metadata2)
	if err := json.Unmarshal(res, &scMetadata); err != nil {
		return nil, parsers.ErrCantContinue{Reason: "can't unmarshal fetched metadata"}
	}
	scMetadata.DownloadURL = fmt.Sprintf("%stracks/%d/download", IE.api2URL, scMetadata.ID)
	return scMetadata, nil
}

func extractInfo(info *metadata2) (*parsers.ExtractorInfo, error) {
	formats, ok := info.getDownloadLink()
	if !ok {
		var err error
		transcodings := info.Media.Transcodings
		formats, err = transcodings.extractFormats()
		if err != nil {
			return nil, fmt.Errorf("can't extract formats from transcodings")
		}
	}

	//duration := float32(info.Duration) * 1 / 1000
	duration, _ := time.ParseDuration(fmt.Sprintf("%dms", info.Duration))
	duration = duration.Round(time.Second)

	thumbnails := extractArtworks(info.ArtworkURL, info.User.AvatarURL)

	var ExtractedInfo = &parsers.ExtractorInfo{
		//ID:        info.ID,
		Permalink: info.Permalink,
		Uploader:  info.User.Username,
		//UploaderID:   info.User.ID,
		//UploaderURL:  info.User.PermalinkURL,
		Timestamp: time.Date(info.CreatedAt.Year(), info.CreatedAt.Month(), info.CreatedAt.Day(), 0, 0, 0, 0, time.UTC),
		Title:     info.Title,
		//Description: info.Description,
		Thumbnails: thumbnails,
		Duration:   duration,
		//WebPageURL:  info.PermalinkURL,
		//License:      info.License,
		//ViewCount:    info.PlaybackCount,
		//LikeCount:    info.LikesCount,
		//CommentCount: info.CommentCount,
		//RepostCount:  info.RepostsCount,
		//Genre:   info.Genre,
		Formats: formats,
	}

	return ExtractedInfo, nil
}

func (info *metadata2) getDownloadLink() (formats parsers.Formats, ok bool) {
	if !info.Downloadable || !info.HasDownloadsLeft {
		return nil, false
	}
	dlURL, err := url.Parse(info.DownloadURL)
	if err != nil {
		// invalid url, just return false
		return nil, false
	}
	res, err := fetch(dlURL)
	if err != nil {
		return nil, false
	}
	var realDlURL struct {
		URL string `json:"redirectUri"`
	}
	if err = json.Unmarshal(res, &realDlURL); err != nil {
		// invalid json, return false
		return nil, false
	}

	format := parsers.Format{
		Url:      realDlURL.URL,
		Ext:      "mp3",
		Type:     "mpeg",
		Protocol: "http",
		Score:    100,
	}
	return []parsers.Format{format}, true
}

func (transcodings transcodings) extractFormats() (parsers.Formats, error) {
	formats := make(parsers.Formats, 0)
	for _, t := range transcodings {
		formatURL, err := url.Parse(t.URL)
		if err != nil {
			// invalid url, continue cycle
			continue
		}
		stream, err := fetch(formatURL)
		if err != nil {
			// can't fetch url, continue cycle
			continue
		}

		var streamObj struct {
			URL string `json:"url"`
		}
		if err = json.Unmarshal(stream, &streamObj); err != nil {
			// invalid json, continue cycle
			continue
		}

		t.URL = streamObj.URL

		formats.Add(t)
	}
	if len(formats) < 1 {
		return nil, fmt.Errorf("not found valid formats inside transodings")
	}
	formats.Sort()
	return formats, nil
}

func extractArtworks(artwork string, avatar string) map[string]parsers.Artwork {
	artworksMap := map[string]int{
		"mini":     16,
		"tiny":     20,
		"small":    32,
		"badge":    47,
		"t67x67":   67,
		"large":    100,
		"t300x300": 300,
		"crop":     400,
		"t500x500": 500,
		"original": 0,
	}
	if len(artwork) < 1 {
		artwork = avatar
	}

	artworks := make(map[string]parsers.Artwork, 0)

	re := regexp.MustCompile(`-([0-9a-z]+)\.jpg`)
	if !re.MatchString(artwork) {
		// no artworks, return empty slice
		return artworks
	}

	for artType, artSize := range artworksMap {
		newType := fmt.Sprintf("-%s.jpg", artType)
		newURL := re.ReplaceAllString(artwork, newType)
		var i = parsers.Artwork{
			Type: artType,
			URL:  newURL,
			Size: artSize,
		}
		artworks[artType] = i
	}

	return artworks
}

func fetch(u *url.URL) ([]byte, error) {
	// loop for two tries
	for i := range []int{0, 0} {
		q := u.Query()
		q.Set("client_id", IE.clientID)
		u.RawQuery = q.Encode()
		res, err := utils.Fetch(u)
		if err != nil {
			// another try
			continue
		}
		if len(res) < 1 && i == 0 {
			// if its first fault, try update token
			if err := updateToken(); err != nil {
				continue
			}
			continue
		}
		return res, nil
	}
	return nil, fmt.Errorf("can't fetch url")
}

func updateToken() error {
	// we are sure in this url, so can skip error
	u, _ := url.Parse("https://soundcloud.com")
	res, err := utils.Fetch(u)
	if err != nil {
		return err
	}
	scriptTmpl := `<script[^>]+src="([^"]+)"`
	clientTmpl := `client_id\s*:\s*"([0-9a-zA-Z]{32})"`
	scriptRE := regexp.MustCompile(scriptTmpl)
	clientRE := regexp.MustCompile(clientTmpl)
	scripts := scriptRE.FindAllSubmatch(res, -1)
	for _, script := range scripts {
		uri := string(script[1])
		scriptURL, err := url.Parse(uri)
		if err != nil {
			// can't parse script url. It's not fatal for us, so just ignore this
			continue
		}
		scriptBody, err := utils.Fetch(scriptURL)
		if err != nil {
			// can't fetch script. It's not fatal for us, so just ignore this
			continue
		}
		matches := clientRE.FindSubmatch(scriptBody)
		if matches == nil {
			continue
		}
		IE.clientID = string(matches[1])
		// just ignore error
		_ = ioutil.WriteFile(tokenFile, matches[1], 0644)
		return nil
	}
	return fmt.Errorf("can't retrieve token")
}
