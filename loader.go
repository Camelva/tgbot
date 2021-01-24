package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"sync"
	"time"
)

type fileInfo struct {
	c              sync.Cond
	cacheContainer *map[int]*fileInfo

	id           int
	fileLocation string
	ttl          time.Duration
}

func (c *Client) downloader() {
	for res := range c.results {
		go c.loadPerUser(res)
	}
}

func (c *Client) loadPerUser(res result) {
	var uid int64
	if res.msg.From != nil {
		uid = int64(res.msg.From.ID)
	} else {
		uid = res.msg.Chat.ID
	}
	user, alreadyActive := c.usersLoading[uid]
	if !alreadyActive {
		user = &sync.WaitGroup{}
		c.usersLoading[uid] = user
	}

	// first waiting if user already doing something
	if alreadyActive {
		log.Println("user already active, waiting")
		c.editMessage(&res.tmpMsg,
			c.getDict(&res.msg).MustLocalize(processQueue))
		user.Wait()
	}

	log.Print(c.loadersInfo())
	log.Printf("adding song from [%d] to queue: [%d] %s", uid, res.song.ID, res.song.Permalink)
	c.capacitor <- struct{}{}

	c.editMessage(&res.tmpMsg,
		c.getDict(&res.msg).MustLocalize(processFetching))

	log.Printf("starting new load per user [%d]", uid)
	user.Add(1)
	c.download(res, uid)
}

func (c *Client) download(res result, uid int64) {
	// check if song already in cache
	info, cached := c.cache[res.song.ID]
	if !cached {
		// if no - add to cache
		info = &fileInfo{
			cacheContainer: &c.cache,
			id:             res.song.ID,
			c:              *sync.NewCond(&sync.RWMutex{}),
			ttl:            c.fileExpiration,
		}
		c.cache[res.song.ID] = info
	}

	info.c.L.Lock()

	var songPath string

	if cached {
		if info.fileLocation == "" {
			info.c.Wait()
		}
		songPath = info.fileLocation
		log.Printf("trying to get from cache: [%d] %s\n", res.song.ID, res.song.Permalink)
	}

	// not cached, need to download and process
	if songPath == "" {
		// can return zero if error occurred
		songPath = c.loadAndPrepareSong(res)
	}

	if songPath != "" {
		c.sendSongToUser(res, songPath)
		log.Printf("done with [%d] %s\n", res.song.ID, res.song.Permalink)
		info.fileLocation = songPath
	}

	info.c.L.Unlock() // unlocking this file
	info.c.Signal()   // telling others song is available
	user, ok := c.usersLoading[uid]
	if ok {
		user.Done()
	}
	<-c.capacitor // telling capacitor we done

	c.removeWhenExpire(info) // scheduling delete from cache and file system
}

func (c *Client) loadAndPrepareSong(res result) string {
	log.Printf("start getting: [%d] %s", res.song.ID, res.song.Permalink)

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
		// until i find a way to send large files
		_ = os.Remove(songPath)
		return ""
	}

	return songPath
}

func (c *Client) sendSongToUser(res result, songPath string) {
	// tell user we good
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
}
