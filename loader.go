package main

import (
	"github.com/camelva/soundcloader"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

type results struct {
	sync.Mutex
	container    map[int64]chan result
	limitPerUser int
	refreshRate  time.Duration
}

func NewResults(limitPerUser int, refreshRate time.Duration) *results {
	return &results{
		container:    make(map[int64]chan result),
		limitPerUser: limitPerUser,
		refreshRate:  refreshRate,
	}
}

func (r *results) Add(c *Client, res result) {
	r.Lock()
	defer r.Unlock()

	userLog := c.log.WithField("userID", res.userID)
	userLog.Info("adding new song for user")

	ch, ok := r.container[res.userID]
	if ok {
		if len(ch) >= r.limitPerUser {
			// too much songs from same user
			userLog.WithField("value", len(ch)).Error("user already sent too much songs")
			c.sendMessage(&res.msg, "please wait for some of your previous songs to download and try again later", true)
			return
		}
		// telling user they need to wait for their previous songs finish
		c.editMessage(&res.tmpMsg, c.getDict(&res.msg).MustLocalize(processQueue))
		ch <- res
		return
	}

	userLog.Debug("first user's song, creating his loader first")

	ch = make(chan result, r.limitPerUser)
	ch <- res
	r.container[res.userID] = ch
	go userLoader(c, res.userID, r.refreshRate)
	return
}

func (r *results) Get(userID int64, log *logrus.Logger) (res result, ok bool) {
	r.Lock()
	defer r.Unlock()

	userLog := log.WithField("userID", userID)
	userLog.Debug("fetching song from user's queue")

	ch, userExist := r.container[userID]
	if userExist {
		if len(ch) > 0 {
			userLog.Debug("fetched some song for user")
			return <-ch, true
		}
		userLog.Debug("got user's queue, but it's empty, deleting")
		delete(r.container, userID)
	}
	userLog.Debug("fetched no songs from user's queue")
	return
}

func (r *results) Len() int {
	return len(r.container)
}

func userLoader(c *Client, userID int64, refreshRate time.Duration) {
	userLog := c.log.WithField("userID", userID)
	tick := time.NewTicker(refreshRate)

	userLog.Debug("starting loader for user")

	for range tick.C {
		res, ok := c.results.Get(userID, c.log)
		if ok {
			userLog.Debug("got new record for user")
			getSong(c, res)
			continue
		}

		userLog.Debug("no more songs for user")
		tick.Stop()
		break
	}
	userLog.Debug("stopping user's loader")
}

type fileInfo struct {
	c              sync.Cond
	cacheContainer *map[int]*fileInfo

	id           int
	fileLocation string
	ttl          time.Duration
}

//func (c *Client) downloader() {
//	c.log.Info("starting downloader")
//	for res := range c.results {
//		userLog := c.log.WithField("userID", res.userID)
//		songLog := userLog.WithFields(logrus.Fields{
//			"id":   res.song.ID,
//			"link": res.song.Permalink,
//		})
//
//		songLog.Info("got new result to load")
//
//		userQueue, ok := c.users[res.userID]
//		if ok {
//			userLog.Debug("this user's loader already up")
//			// means loader *should be* already up
//			if len(userQueue) == cap(userQueue) {
//				userLog.WithField("amount", len(userQueue)).Info("user already sent too much songs")
//				c.sendMessage(
//					&res.msg,
//					"please wait for your previous songs to download and try again later",
//					true,
//				)
//				return
//			}
//			c.editMessage(&res.tmpMsg, c.getDict(&res.msg).MustLocalize(processQueue))
//			userQueue <- res
//
//			continue
//		}
//		userLog.Debug("need to create user instance")
//		// create user queue
//		userQueue = make(chan result, 30)
//		userQueue <- res
//		c.users[res.userID] = userQueue
//
//		go freeUser(res.userID, c.users, userLog)
//
//		go c._loadUserQueue(res.userID, c.users[res.userID])
//	}
//}

//func freeUser(uid int64, users map[int64]chan result, log *logrus.Entry) {
//	ticker := time.NewTicker(time.Minute * 30)
//	log.Debug("start ticker to free user")
//	defer ticker.Stop()
//	for range ticker.C {
//		log.Debug("tick")
//		if len(users[uid]) == 0 {
//			log.Debug("user's channel is empty, freeing")
//			// no more entries, free user
//			close(users[uid])
//			break
//		}
//		continue
//	}
//}

//func (c *Client) _loadUserQueue(uid int64, in chan result) {
//	userLog := c.log.WithField("userID", uid)
//	userLog.Debug("init user loader")
//	for res := range in {
//		songLog := userLog.WithFields(logrus.Fields{
//			"id":   res.song.ID,
//			"link": res.song.Permalink,
//		})
//		songLog.Info("got new song for user")
//		// adding record to limiter, if already max - waiting for free space
//		c.capacitor.Add(res.song.ID, res.song.Permalink)
//		songLog.Debug("added to capacitor, start getting")
//		c._load(res)
//	}
//	userLog.Trace("no more songs from user")
//	delete(c.users, uid)
//}

func getSong(c *Client, res result) {
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
