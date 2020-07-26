package parsers

import (
	"net/url"
	"regexp"
	"sort"
	"time"
)

type Extractor interface {
	Name() string
	Extract(url.URL) (*ExtractorInfo, error)
	Compatible(url.URL) bool
}

type Format struct {
	Url      string
	Ext      string
	Type     string
	Protocol string
	Score    int
}

type Formats []Format

func (formats *Formats) Add(t Transcodinger) {
	re := regexp.MustCompile(`_`)
	ext := re.Split(t.GetPreset(), -1)[0]
	re = regexp.MustCompile(`audio/([\w-]+)[;]?`)
	mimeType := re.FindStringSubmatch(t.GetMimeType())[1]
	f := Format{
		Url:      t.GetURL(),
		Type:     mimeType,
		Protocol: t.GetProtocol(),
		Ext:      ext,
	}
	*formats = append(*formats, f)
}
func (formats *Formats) Sort() {
	formatsCopy := make(Formats, len(*formats))
	copy(formatsCopy, *formats)
	for i, format := range formatsCopy {
		var score int
		switch format.Ext {
		case "mp3":
			score += 10
		case "opus":
			score += 5
		default:
			score += 0
		}
		switch format.Protocol {
		case "progressive":
			score += 10
		case "hls":
			score += 5
		default:
			score += 0
		}
		formatsCopy[i].Score = score
	}
	sort.Slice(formatsCopy, func(i, j int) bool { return formatsCopy[i].Score > formatsCopy[j].Score })
	*formats = formatsCopy
}

type ExtractorInfo struct {
	ID        int
	Permalink string
	Uploader  string
	//UploaderID   int
	//UploaderURL  string
	Timestamp   time.Time
	Title       string
	Description string
	Thumbnails  map[string]Artwork
	Duration    float32
	WebPageURL  string
	//License      string
	//ViewCount    int
	//LikeCount    int
	//CommentCount int
	//RepostCount  int
	Genre   string
	Formats Formats
}

type Transcodinger interface {
	GetURL() string
	GetPreset() string
	GetProtocol() string
	GetMimeType() string
}

type Artwork struct {
	Type string
	URL  string
	Size int
}
