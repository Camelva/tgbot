package main

import (
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

//func (c *Client) downloader() {
//	for res := range c.results {
//		c.log.WithFields(logrus.Fields{
//			"id":   res.song.ID,
//			"link": res.song.Permalink,
//		}).Trace("got new result to load")
//
//		go c.loadPerUser(res)
//	}
//}
//
//func (c *Client) loadPerUser(res result) {
//	var uid int64
//	if res.msg.From != nil {
//		uid = int64(res.msg.From.ID)
//	} else {
//		uid = res.msg.Chat.ID
//	}
//
//	userLog := c.log.WithField("userID", uid)
//
//	user, alreadyActive := c.usersLoading[uid]
//	if !alreadyActive {
//		userLog.Trace("user not active yet, adding..")
//		user = &sync.WaitGroup{}
//		c.usersLoading[uid] = user
//	}
//
//	// first waiting if user already doing something
//	if alreadyActive {
//		userLog.Trace("user already active, waiting..")
//		c.editMessage(&res.tmpMsg,
//			c.getDict(&res.msg).MustLocalize(processQueue))
//		user.Wait()
//		userLog.Trace("user no more active, working..")
//	}
//
//	c.log.Trace(c.loadersInfo())
//	songLog := userLog.WithFields(logrus.Fields{
//		"id":   res.song.ID,
//		"link": res.song.Permalink,
//	})
//
//	songLog.Trace("adding song to queue, waiting..")
//	c.capacitor <- struct{}{}
//	songLog.Trace("queue is open, working..")
//
//	c.editMessage(&res.tmpMsg,
//		c.getDict(&res.msg).MustLocalize(processFetching))
//
//	user.Add(1)
//	c.download(res, uid)
//	done <- struct{}{}
//}
//
//func (c *Client) download(res result, uid int64) {
//	songLog := c.log.WithFields(logrus.Fields{
//		"userID": uid,
//		"id":     res.song.ID,
//		"link":   res.song.Permalink,
//	})
//	// check if song already in cache
//	info, cached := c.cache[res.song.ID]
//	if !cached {
//		songLog.Trace("not in cache, creating record..")
//		// if no - add to cache
//		info = &fileInfo{
//			cacheContainer: &c.cache,
//			id:             res.song.ID,
//			c:              *sync.NewCond(&sync.RWMutex{}),
//			ttl:            c.fileExpiration,
//		}
//		c.cache[res.song.ID] = info
//	}
//
//	songLog.Trace("locking this file..")
//	info.c.L.Lock()
//	songLog.Trace("file locked for us, working..")
//
//	var songPath string
//
//	if cached {
//		if info.fileLocation == "" {
//			songLog.Trace("file cached, but location is still empty, waiting")
//			info.c.Wait()
//			songLog.Trace("continue work with file..")
//		}
//		songPath = info.fileLocation
//		songLog.Trace("getting song from cache")
//	}
//
//	// not cached, need to download and process
//	if songPath == "" {
//		// can return zero if error occurred
//		songLog.Trace("not cached, need to download..")
//		songPath = c.loadAndPrepareSong(res)
//	}
//
//	if songPath != "" {
//		songLog.Trace("trying to send song to user..")
//		c.sendSongToUser(res, songPath)
//		songLog.Info("done with this song")
//		info.fileLocation = songPath
//	}
//
//	songLog.Trace("unlocking file..")
//	info.c.L.Unlock() // unlocking this file
//
//	songLog.Trace("telling others it is available now..")
//	info.c.Signal() // telling others song is available
//
//	user, ok := c.usersLoading[uid]
//	if ok {
//		songLog.Trace("allowing user to load next song..")
//		user.Done()
//	}
//	songLog.Trace("freeing capacitor by one..")
//	<-c.capacitor // telling capacitor we done
//
//	c.removeWhenExpire(info) // scheduling delete from cache and file system
//}
//
//func (c *Client) loadAndPrepareSong(res result) string {
//	songLog := c.log.WithFields(logrus.Fields{
//		"id":   res.song.ID,
//		"link": res.song.Permalink,
//	})
//
//	songLog.Trace("getting song by using lib..")
//	songPath, err := res.song.GetNext()
//
//	if err != nil {
//		songLog.WithError(err).Error("can't get song")
//		c.sendMessage(&res.msg, c.getDict(&res.msg).MustLocalize(errUndefined(err)), true)
//		return ""
//	}
//
//	fileStats, err := os.Stat(songPath)
//	if err != nil {
//		songLog.WithFields(logrus.Fields{
//			"error": err,
//			"path":  songPath,
//		}).Error("can't read file..")
//		c.sendMessage(&res.msg, c.getDict(&res.msg).MustLocalize(errUndefined(err)), true)
//		return ""
//	}
//	if fileStats.Size() > (49 * 1024 * 1024) {
//		songLog.WithField("size", fileStats.Size()/1024/1024).Error("file >50mb")
//		c.sendMessage(&res.msg, c.getDict(&res.msg).MustLocalize(errSizeLimit), true)
//
//		songLog.Trace("removing file..")
//		_ = os.Remove(songPath)
//		return ""
//	}
//
//	return songPath
//}
//
//func (c *Client) sendSongToUser(res result, songPath string) {
//	songLog := c.log.WithFields(logrus.Fields{
//		"id":   res.song.ID,
//		"link": res.song.Permalink,
//	})
//	songLog.WithField("file", songPath).Trace("preparing file to loading..")
//	// tell user we good
//	c.editMessage(&res.tmpMsg,
//		c.getDict(&res.msg).MustLocalize(processUploading))
//
//	audioMsg := tgbotapi.NewAudioUpload(res.msg.Chat.ID, songPath)
//	audioMsg.Title = res.song.Title
//	audioMsg.Performer = res.song.Author
//	audioMsg.Duration = int(res.song.Duration.Seconds())
//	audioMsg.ReplyToMessageID = res.msg.MessageID
//
//	if _, err := c.bot.Send(audioMsg); err != nil {
//		songLog.WithError(err).Error("can't send song to user")
//		c.sendMessage(&res.msg, c.getDict(&res.msg).MustLocalize(errUndefined(err)), true)
//	}
//
//	c.deleteMessage(&res.tmpMsg)
//}

func (c *Client) downloader() {
	c.log.Trace("starting downloader")
	for res := range c.results {
		var uid int64
		if res.msg.From != nil {
			uid = int64(res.msg.From.ID)
		} else {
			uid = res.msg.Chat.ID
		}
		userLog := logrus.WithField("userID", uid)
		songLog := userLog.WithFields(logrus.Fields{
			"id":   res.song.ID,
			"link": res.song.Permalink,
		})

		songLog.Trace("got new result to load")

		userQueue, ok := c.users[uid]
		if ok {
			userLog.Trace("user loader already up")
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
			c.editMessage(&res.msg, c.getDict(&res.msg).MustLocalize(processQueue))
			userQueue <- res

			continue
		}
		userLog.Trace("need to create user instance")
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
	log.Trace("start ticker to free user")
	defer ticker.Stop()
	for tick := range ticker.C {
		log.WithField("tick", tick).Trace("tick")
		if len(users[uid]) == 0 {
			log.Trace("channel empty, freeing")
			// no more entries, free user
			close(users[uid])
			break
		}
		continue
	}
}

func (c *Client) _loadUserQueue(uid int64, in chan result) {
	userLog := c.log.WithField("userID", uid)
	userLog.Trace("init user loader")
	for res := range in {
		userLog.Trace("got song for user")
		// adding record to limiter, if already max - waiting for free space
		c.capacitor <- struct{}{}
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

	c.editMessage(&res.msg, c.getDict(&res.msg).MustLocalize(processFetching))

	// check if song already in cache
	info, cached := c.cache[res.song.ID]
	if !cached {
		songLog.Trace("not in cache, creating record..")
		// if no - add to cache
		info = &fileInfo{
			cacheContainer: &c.cache,
			id:             res.song.ID,
			c:              *sync.NewCond(&sync.Mutex{}),
			ttl:            c.fileExpiration,
		}
		c.cache[res.song.ID] = info
	}

	songLog.Trace("locking this file..")
	info.c.L.Lock()
	songLog.Trace("file locked only for us, working..")

	var songPath string

	if cached {
		if info.fileLocation == "" {
			songLog.Trace("file cached, but location is still empty, waiting")
			info.c.Wait()
			songLog.Trace("continue work with file..")
		}
		songPath = info.fileLocation
		songLog.Trace("getting song from cache")
	}

	// not cached, need to download and process
	if songPath == "" {
		// can return zero if error occurred
		songLog.Trace("not cached, need to download..")
		songPath = c._loadAndPrepareSong(res)
	}

	if songPath != "" {
		songLog.Trace("trying to send song to user..")
		c._sendSongToUser(res, songPath)
		songLog.Info("done with this song")
		info.fileLocation = songPath
	}

	c.deleteMessage(&res.tmpMsg)

	songLog.Trace("unlocking file..")
	info.c.L.Unlock() // unlocking this file

	songLog.Trace("telling others it is available now..")
	info.c.Signal() // telling others song is available

	songLog.Trace("freeing capacitor by one..")
	<-c.capacitor // telling capacitor we done

	c.removeWhenExpire(info) // scheduling delete from cache and file system
}

func (c *Client) _loadAndPrepareSong(res result) string {
	songLog := c.log.WithFields(logrus.Fields{
		"id":   res.song.ID,
		"link": res.song.Permalink,
	})

	songLog.Trace("getting song by using lib..")
	songPath, err := res.song.GetNext()

	if err != nil {
		songLog.WithError(err).Error("can't get song")
		c.sendMessage(&res.msg, c.getDict(&res.msg).MustLocalize(errUndefined(err)), true)
		return ""
	}

	fileStats, err := os.Stat(songPath)
	if err != nil {
		songLog.WithFields(logrus.Fields{
			"error": err,
			"path":  songPath,
		}).Error("can't read file..")
		c.sendMessage(&res.msg, c.getDict(&res.msg).MustLocalize(errUndefined(err)), true)
		return ""
	}
	if fileStats.Size() > (49 * 1024 * 1024) {
		songLog.WithField("size", fileStats.Size()/1024/1024).Error("file >50mb")
		c.sendMessage(&res.msg, c.getDict(&res.msg).MustLocalize(errSizeLimit), true)

		songLog.Trace("removing file..")
		_ = os.Remove(songPath)
		return ""
	}

	return songPath
}

func (c *Client) _sendSongToUser(res result, songPath string) {
	songLog := c.log.WithFields(logrus.Fields{
		"id":   res.song.ID,
		"link": res.song.Permalink,
	})
	songLog.WithField("file", songPath).Trace("preparing file to loading..")

	// tell user we good
	c.editMessage(&res.tmpMsg,
		c.getDict(&res.msg).MustLocalize(processUploading))

	audioMsg := tgbotapi.NewAudioUpload(res.msg.Chat.ID, songPath)
	audioMsg.Title = res.song.Title
	audioMsg.Performer = res.song.Author
	audioMsg.Duration = int(res.song.Duration.Seconds())
	audioMsg.ReplyToMessageID = res.msg.MessageID

	if _, err := c.bot.Send(audioMsg); err != nil {
		songLog.WithError(err).Error("can't send song to user")
		c.sendMessage(&res.msg, c.getDict(&res.msg).MustLocalize(errUndefined(err)), true)
	}
}
