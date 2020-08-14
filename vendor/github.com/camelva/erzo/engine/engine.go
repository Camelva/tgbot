package engine

import (
	"fmt"
	"github.com/camelva/erzo/loaders"
	"github.com/camelva/erzo/parsers"
	"github.com/camelva/erzo/utils"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"time"
)

var _extractors = map[string]parsers.Extractor{}
var _loaders = map[string]loaders.Loader{}

const (
	_urlPattern = `((?:[a-z]{3,6}:\/\/)|(?:^|\s))` +
		`((?:[a-zA-Z0-9\-]+\.)+[a-z]{2,13})` +
		`([\.\?\=\&\%\/\w\-]*\b)`
)

type SongResult struct {
	Path       string
	Author     string
	Title      string
	Thumbnails map[string]parsers.Artwork
	Duration   time.Duration
	UploadDate time.Time
}

type SongInfo struct {
	Info     *parsers.ExtractorInfo
	Metadata []string
	URL      string

	engine *Engine
}

func AddExtractor(x parsers.Extractor) {
	name := x.Name()
	_extractors[name] = x
	//_extractors = append(_extractors, x)
}
func Extractors() map[string]parsers.Extractor {
	return _extractors
}

func AddLoader(l loaders.Loader) {
	name := l.Name()
	_loaders[name] = l
	//_loaders = append(_loaders, l)
}
func Loaders() map[string]loaders.Loader {
	return _loaders
}

type Engine struct {
	extractors   map[string]parsers.Extractor
	loaders      map[string]loaders.Loader
	outputFolder string
}

// New return new instance of Engine
func New(out string, truncate bool) *Engine {
	xtrs := Extractors()
	ldrs := Loaders()
	if (len(xtrs) < 1) || (len(ldrs) < 1) {
		// we need at least 1 extractor and 1 loader for work
		return nil
	}
	e := &Engine{
		extractors:   xtrs,
		loaders:      ldrs,
		outputFolder: out,
	}
	if truncate {
		e.Clean()
	}
	return e
}

// Clean current e.OutputFolder directory
func (e Engine) Clean() {
	_ = os.RemoveAll(e.outputFolder)
	return
}

// Process your message. Return file name or one of this errors:
// ErrNotURL if there is no urls in your message
// ErrUnsupportedService if url belongs to unsupported service
// ErrUnsupportedType if service supported but certain type - not yet
// ErrCantFetchInfo if fatal error occurred while extracting info from url
// ErrUnsupportedProtocol if there is no downloader for this format
// ErrDownloadingError if fatal error occurred while downloading song
// ErrUndefined any other errors
func (e Engine) Process(s string) (*SongResult, error) {
	song, err := e.GetInfo(s)
	if err != nil {
		return nil, err
	}
	return song.Get()
}

func (e Engine) GetInfo(s string) (*SongInfo, error) {
	u, ok := ExtractURL(s)
	if !ok {
		return nil, ErrNotURL{}
	}
	info, err := e.extractInfo(*u)
	if err != nil {
		return nil, err
	}
	return &SongInfo{
		Info:   info,
		URL:    info.Formats[0].Url,
		engine: &e,
	}, nil
}

func (s *SongInfo) Get() (*SongResult, error) {
	s.Metadata = createMetadata(s.Info)
	filePath, err := s.engine.downloadSong(s.Info, s.Metadata)
	if err != nil {
		return nil, err
	}
	return &SongResult{
		Path:       filePath,
		Author:     s.Info.Uploader,
		Title:      s.Info.Title,
		Thumbnails: s.Info.Thumbnails,
		Duration:   s.Info.Duration,
		UploadDate: s.Info.Timestamp,
	}, nil
}

func (e Engine) extractInfo(u url.URL) (*parsers.ExtractorInfo, error) {
	for _, xtr := range e.extractors {
		if !xtr.Compatible(u) {
			continue
		}
		info, err := xtr.Extract(u)
		if err != nil {
			switch err.(type) {
			case parsers.ErrFormatNotSupported:
				return nil, ErrUnsupportedType{err.(parsers.ErrFormatNotSupported)}
			case parsers.ErrCantContinue:
				return nil, ErrDownloadingError{err.Error()}
			default:
				return nil, ErrUndefined{}
			}
		}
		return info, nil
	}
	return nil, ErrUnsupportedService{Service: u.Hostname()}
}

func (e Engine) downloadSong(info *parsers.ExtractorInfo, metadata []string) (string, error) {
	if _, err := ioutil.ReadDir(e.outputFolder); err != nil {
		// outputFolder don't exist. Creating it...
		if err := os.Mkdir(e.outputFolder, 0700); err != nil {
			// can't create outPutFolder. Going to save files in root directory
			e.outputFolder = ""
		}
	}
	outPath := makeFilePath(e.outputFolder, info.Permalink)
	imageURL, err := url.Parse(info.Thumbnails["original"].URL)
	var thumbnail string
	if err == nil {
		res, err := utils.Fetch(imageURL)
		if err == nil {
			thumbnail = path.Join(e.outputFolder, imageURL.Path)
			if err := ioutil.WriteFile(thumbnail, res, 0644); err != nil {
				thumbnail = ""
			}
		}
	}
	var downloadingErr error
	for _, format := range info.Formats {
		u, err := url.Parse(format.Url)
		if err != nil {
			// invalid url, try another
			continue
		}
		for _, ldr := range e.loaders {
			if !ldr.Compatible(format) {
				// incompatible with loader, try another one
				continue
			}
			if err := ldr.Get(u, outPath); err != nil {
				// save err
				downloadingErr = err
				continue
			}
			if err := ldr.UpdateTags(outPath, metadata); err != nil {
				log.Println(err)
			}
			if len(thumbnail) > 0 {
				if err := ldr.AddThumbnail(outPath, thumbnail); err != nil {
					log.Println(err)
				}
				_ = os.Remove(thumbnail)
			}
			return outPath, nil
		}
	}
	if downloadingErr != nil {
		return "", ErrDownloadingError{Reason: downloadingErr.Error()}
	}
	return "", ErrUnsupportedProtocol{}
}

func createMetadata(info *parsers.ExtractorInfo) []string {
	metaMap := map[string]string{
		"title": info.Title,
		"album": info.Title,
		//"genre":      info.Genre,
		"artist":       info.Uploader,
		"album_artist": info.Uploader,
		"track":        strconv.Itoa(1),
		"date":         strconv.Itoa(info.Timestamp.Year()),
	}
	var metadata = make([]string, 0, len(metaMap))
	for key, value := range metaMap {
		line := fmt.Sprintf("%s=%s", key, value)
		metadata = append(metadata, line)
	}
	return metadata
}

func makeFilePath(folder string, title string) string {
	fileName := fmt.Sprintf("%s.mp3", title)
	outPath := path.Join(folder, fileName)
	//if _, err := ioutil.ReadFile(outPath); err == nil {
	//	title = fmt.Sprintf("%s-copy", title)
	//	return makeFilePath(folder, title)
	//}
	return outPath
}

// ExtractURL trying to extract url from message
func ExtractURL(message string) (u *url.URL, ok bool) {
	re := regexp.MustCompile(_urlPattern)
	rawURL := re.FindString(message)
	if len(rawURL) < 1 {
		return nil, false
	}
	link, err := url.Parse(rawURL)
	if err != nil {
		return nil, false
	}
	return link, true
}
