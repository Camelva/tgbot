package main

import (
	"github.com/camelva/soundcloader"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
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
	c.log.Info("starting downloader")
	for res := range c.results {
		var uid int64
		if res.msg.From != nil {
			uid = int64(res.msg.From.ID)
		} else {
			uid = res.msg.Chat.ID
		}
		userLog := c.log.WithField("userID", uid)
		songLog := userLog.WithFields(logrus.Fields{
			"id":   res.song.ID,
			"link": res.song.Permalink,
		})

		songLog.Info("got new result to load")

		userQueue, ok := c.users[uid]
		if ok {
			userLog.Debug("this user's loader already up")
			// means loader *should be* already up
			if len(userQueue) == cap(userQueue) {
				userLog.WithField("amount", len(userQueue)).Info("user already sent too much songs")
				c.sendMessage(
					&res.msg,
					"please wait for your previous songs to download and try again later",
					true,
				)
				return
			}
			c.editMessage(&res.tmpMsg, c.getDict(&res.msg).MustLocalize(processQueue))
			userQueue <- res

			continue
		}
		userLog.Debug("need to create user instance")
		// create user queue
		userQueue = make(chan result, 30)
		userQueue <- res
		c.users[uid] = userQueue

		go freeUser(uid, c.users, userLog)

		go c._loadUserQueue(uid, c.users[uid])
	}
}

func freeUser(uid int64, users map[int64]chan result, log *logrus.Entry) {
	ticker := time.NewTicker(time.Minute * 30)
	log.Debug("start ticker to free user")
	defer ticker.Stop()
	for range ticker.C {
		log.Debug("tick")
		if len(users[uid]) == 0 {
			log.Debug("user's channel is empty, freeing")
			// no more entries, free user
			close(users[uid])
			break
		}
		continue
	}
}

func (c *Client) _loadUserQueue(uid int64, in chan result) {
	userLog := c.log.WithField("userID", uid)
	userLog.Debug("init user loader")
	for res := range in {
		songLog := userLog.WithFields(logrus.Fields{
			"id":   res.song.ID,
			"link": res.song.Permalink,
		})
		songLog.Info("got new song for user")
		// adding record to limiter, if already max - waiting for free space
		c.capacitor.Add(res.song.ID, res.song.Permalink)
		songLog.Debug("added to capacitor, start getting")
		c._load(res)
	}
	userLog.Trace("no more songs from user")
	delete(c.users, uid)
}

func (c *Client) _load(res result) {
	songLog := c.log.WithFields(logrus.Fields{
		"id":   res.song.ID,
		"link": res.song.Permalink,
	})

	c.editMessage(&res.tmpMsg, c.getDict(&res.msg).MustLocalize(processFetching))

	// check if song already in cache
	info, cached := c.cache[res.song.ID]
	if !cached {
		songLog.Debug("not in cache, creating record..")
		// if no - add to cache
		info = &fileInfo{
			cacheContainer: &c.cache,
			id:             res.song.ID,
			c:              *sync.NewCond(&sync.Mutex{}),
			ttl:            c.fileExpiration,
		}
		c.cache[res.song.ID] = info
	}

	songLog.Debug("locking this file..")
	info.c.L.Lock()
	songLog.Debug("file locked only for us, working..")

	defer func() {
		// regardless of result:
		c.deleteMessage(&res.tmpMsg) // delete temp message

		songLog.Debug("unlocking file..")
		info.c.L.Unlock() // unlock this file

		//songLog.Trace("telling others it is available now..")
		//info.c.Signal() // tell others this song is available

		songLog.Debug("freeing capacitor by one..")
		c.capacitor.Remove(res.song.ID) // free space for next song

		c.removeWhenExpire(info) // scheduling delete from cache and file system
	}()

	var songPath string

	// because we locking mutex, situation when fileLocation field *yet* empty is impossible
	// empty field from cache only means something went wrong earlier
	if info.fileLocation != "" {
		songLog.Debug("trying to get song from cache")
		songPath = info.fileLocation
	}

	// not cached, need to download and process
	if songPath == "" {
		// can return zero if error occurred
		songLog.Debug("not cached, need to download..")
		songPath = c._loadAndPrepareSong(res, songLog)
	}

	if songPath == "" {
		// means this song is unavailable
		// remove entry from cache and return
		delete(c.cache, res.song.ID)
		return
	}

	songLog.Debug("trying to send song to user..")
	c._sendSongToUser(res, songPath, songLog)

	songLog.Info("done with this song")
	info.fileLocation = songPath
}

func (c *Client) _loadAndPrepareSong(res result, songLog *logrus.Entry) string {
	songLog.Debug("getting song by using lib..")
	songPath, err := res.song.GetNext()

	if err != nil {
		songLog.WithError(err).Error("can't get song")

		if err == soundcloader.EmptyStream {
			c.sendMessage(&res.msg, c.getDict(&res.msg).MustLocalize(errUnavailableSong), true)
		} else {
			c.sendMessage(&res.msg, c.getDict(&res.msg).MustLocalize(errUndefined(err)), true)
		}

		return ""
	}

	fileStats, err := os.Stat(songPath)
	if err != nil {
		songLog.WithField("value", songPath).WithError(err).Error("can't read file..")
		c.sendMessage(&res.msg, c.getDict(&res.msg).MustLocalize(errUndefined(err)), true)
		return ""
	}
	if fileStats.Size() > (49 * 1024 * 1024) {
		songLog.WithField("value", fileStats.Size()/1024/1024).Error("file >50mb")
		c.sendMessage(&res.msg, c.getDict(&res.msg).MustLocalize(errSizeLimit), true)

		songLog.Debug("removing file..")
		_ = os.Remove(songPath)
		return ""
	}

	return songPath
}

func (c *Client) _sendSongToUser(res result, songPath string, songLog *logrus.Entry) {
	songLog.WithField("value", songPath).Debug("preparing file to loading..")

	// tell user we good
	c.editMessage(&res.tmpMsg,
		c.getDict(&res.msg).MustLocalize(processUploading))

	audioMsg := tgbotapi.NewAudioUpload(res.msg.Chat.ID, songPath)
	audioMsg.Title = res.song.Title
	audioMsg.Performer = res.song.Author
	audioMsg.Duration = int(res.song.Duration.Seconds())
	audioMsg.ReplyToMessageID = res.msg.MessageID

	audioMsg.Caption = "@scdl_info"

	if _, err := c.bot.Send(audioMsg); err != nil {
		songLog.WithError(err).Error("can't send song to user")
		c.sendMessage(&res.msg, c.getDict(&res.msg).MustLocalize(errUndefined(err)), true)
	}
}
