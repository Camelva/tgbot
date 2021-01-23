package soundcloader

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Stream struct {
	parent *Song

	Format,
	Extension,
	URL,
	Description string
}

func (s *Stream) Get() (fileLocation string, err error) {
	// make sure output folder exist
	err = os.MkdirAll(s.parent.client.OutputFolder, 0755)
	if err != nil {
		return "", fmt.Errorf("cant access output folder: %s", err)
	}

	if s.Format != "original" {
		filename := fmt.Sprintf("%s.%s", s.parent.Permalink, s.Extension)
		fileLocation = filepath.Join(s.parent.client.OutputFolder, filename)
		if alreadyExist(fileLocation) {
			return fileLocation, nil
		}
	}

	if s.URL == "" {
		if s.parent.client.Debug {
			log.Printf("Current stream URL empty\n")
		}
		return "", EmptyStream
	}

	res, err := s.parent.client.fetch(s.URL, true)
	if err != nil {
		if s.parent.client.Debug {
			log.Printf("Can't fetch url: %s\n", err)
		}
		return
	}

	wrappedRes := new(wrappedURL)
	if err = json.Unmarshal(res, wrappedRes); err != nil {
		if s.parent.client.Debug {
			log.Printf("Invalid json: %s\n", wrappedRes)
		}
		return "", InvalidJSON
	}

	if s.Format == "hls" {
		return s.parent.client.ffmpegLoad(
			fileLocation,
			oneOf(wrappedRes.URL, wrappedRes.RedirectURI),
		)
	}
	var useOriginalName = false
	if s.Format == "original" {
		useOriginalName = true
	}

	return s.parent.client.nativeLoad(
		fileLocation,
		oneOf(wrappedRes.URL, wrappedRes.RedirectURI),
		useOriginalName,
	)
}

func sortStreams(ss []Stream) []Stream {
	newSlice := make([]Stream, 0, len(ss))
	var mp3Progressive, mp3HLS, opusHSL, original Stream
	for _, s := range ss {
		switch s.Description {
		case "original":
			original = s
		case "mp3-progressive":
			mp3Progressive = s
		case "mp3-hls":
			mp3HLS = s
		case "opus-hls":
			opusHSL = s
		}
	}
	newSlice = append(newSlice, mp3Progressive, mp3HLS, opusHSL)
	if (original != Stream{}) {
		newSlice = append(newSlice, original)
	}
	return newSlice
}
