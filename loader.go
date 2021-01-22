package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"sync"
)

type fileState struct {
	mu     sync.Mutex
	active int

	filePath string
}

func (c *Client) downloader(sourcesChan chan result) {
	var sameTimeLoadersLimit = 10
	c.workingLoaders = make(map[string]*fileState)
	c.capacitor = make(chan struct{}, sameTimeLoadersLimit)

	//var workingGoroutines = map[string]*fileState{}
	for res := range sourcesChan {
		log.Print(c.loadersInfo())
		log.Printf("adding loader to queue: %s", res.song.ID)

		c.editMessage(&res.tmpMsg,
			c.getDict(&res.msg).MustLocalize(processFetching))

		c.capacitor <- struct{}{}
		go c.download(res)
	}
}

func (c *Client) download(res result) {
	var pool = c.workingLoaders

	state := new(fileState)
	if st, ok := pool[res.song.ID]; ok {
		// work for this song already exist, taking their state
		state = st
	} else {
		pool[res.song.ID] = state
	}

	state.active++
	state.mu.Lock()

	var songPath string

	if state.filePath == "" {
		r := c.loadAndPrepareSong(res)
		if r == "" {
			return
		}

		songPath = r
	} else {
		songPath = state.filePath
	}

	c.editMessage(&res.tmpMsg,
		c.getDict(&res.msg).MustLocalize(processUploading))

	audioMsg := tgbotapi.NewAudioUpload(res.msg.Chat.ID, songPath)
	audioMsg.Title = res.song.Title
	audioMsg.Performer = res.song.Author
	audioMsg.Duration = int(res.song.Duration.Seconds())
	audioMsg.ReplyToMessageID = res.msg.MessageID

	if _, err := c.bot.Send(audioMsg); err != nil {
		log.Printf("can't send song to user: %s\n", err)
		c.sendMessage(&res.msg, c.getDict(&res.msg).MustLocalize(errUndefined(err)), true)
	}

	c.deleteMessage(&res.tmpMsg)

	log.Printf("done with %s", res.song.ID)

	state.active--

	state.filePath = songPath
	state.mu.Unlock()
	<-c.capacitor

	// no one waiting with same filename, removing file and record from pool
	if state.active < 1 {
		_ = os.Remove(songPath)
		delete(pool, res.song.ID)
	}
}

func (c *Client) loadAndPrepareSong(res result) string {
	log.Printf("start getting: [%s]\n", res.song.ID)

	songPath, err := res.song.GetNext()

	if err != nil {
		log.Printf("error with getting song: %s\n", err)
		c.sendMessage(&res.msg, c.getDict(&res.msg).MustLocalize(errUndefined(err)), true)
		return ""
	}
	fileStats, err := os.Stat(songPath)
	if err != nil {
		log.Printf("error with reading file: %s\n", err)
		c.sendMessage(&res.msg, c.getDict(&res.msg).MustLocalize(errUndefined(err)), true)
		return ""
	}
	if fileStats.Size() > (49 * 1024 * 1024) {
		log.Printf("file exceed 50mb\n")
		c.sendMessage(&res.msg, c.getDict(&res.msg).MustLocalize(errSizeLimit), true)
		return ""
	}

	return songPath
}
