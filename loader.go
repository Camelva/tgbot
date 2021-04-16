package main

import (
	"errors"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/camelva/soundcloader"
	"os"
	"path/filepath"
	"strconv"
	"tgbot/resp"
	"tgbot/storage"
)

// max allowed size is 50Mb
var MaxSize int64 = 50 * 1024 * 1024

func loadSong(b *gotgbot.Bot, ctx *ext.Context) (err error) {
	tempMessage, ok := ctx.Data["tempMessage"].(*gotgbot.Message)
	if !ok {
		e := errors.New("no tempMessage")
		log.WithError(e).Error("internal error")

		_, _ = b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrUndefined(e), ctx.EffectiveUser.LanguageCode),
			&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
		return e
	}
	delete(ctx.Data, "tempMessage")

	parsedURL, ok := ctx.Data["parsedURL"].(*soundcloader.URLInfo)
	if !ok {
		e := errors.New("no parsedURL")
		log.WithError(e).Error("internal error")

		_, _ = b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrUndefined(e), ctx.EffectiveUser.LanguageCode),
			&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
		return e
	}
	delete(ctx.Data, "parsedURL")

	defer func() {
		_, _ = tempMessage.Delete(b)
	}()

	tempMessage, err = tempMessage.EditText(b, resp.Get(resp.ProcessFetching, ctx.EffectiveUser.LanguageCode),
		&gotgbot.EditMessageTextOpts{ParseMode: "HTML"})
	if err != nil {
		return err
	}

	// making separate folder for each message to avoid conflicts
	sClient := *soundcloader.DefaultClient
	sClient.OutputFolder = filepath.Join(sClient.OutputFolder, strconv.FormatInt(ctx.UpdateId, 10))

	defer os.RemoveAll(sClient.OutputFolder)

	localLog := log.
		WithField("value", parsedURL.String()).
		WithField("messageID", ctx.EffectiveMessage.MessageId)

	song, e := sClient.GetURL(parsedURL)
	if e != nil {
		if e == soundcloader.NotSong {
			localLog.Info("not song url, exiting..")
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrUnsupportedFormat, ctx.EffectiveUser.LanguageCode),
				&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
		} else {
			localLog.WithError(e).Error("can't get url")
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrUndefined(e), ctx.EffectiveUser.LanguageCode),
				&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
		}
		return err
	}

	location, err := song.GetNext()
	if err != nil {
		if err == soundcloader.EmptyStream {
			localLog.Error("empty stream")
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrUnavailableSong, ctx.EffectiveUser.LanguageCode),
				&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId})
		} else {
			localLog.WithError(err).Error("undefined error")
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrUndefined(err), ctx.EffectiveUser.LanguageCode),
				&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId})
		}
		return err
	}

	stats, err := os.Stat(location)
	if err != nil {
		localLog.WithError(err).Error("file stat error")
		_, _ = b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrUndefined(err), ctx.EffectiveUser.LanguageCode),
			&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId})
		return err
	}

	ctx.Data["fileLocation"] = location

	if stats.Size() >= MaxSize {
		localLog.Info("file size limit, trying external storage")
		tempMessage, _ = tempMessage.EditText(b,
			resp.Get(resp.ProcessStorage, ctx.EffectiveUser.LanguageCode),
			&gotgbot.EditMessageTextOpts{ParseMode: "HTML"})

		return externalUpload(b, ctx)
	}

	tempMessage, err = tempMessage.EditText(b,
		resp.Get(resp.ProcessUploading, ctx.EffectiveUser.LanguageCode),
		&gotgbot.EditMessageTextOpts{ParseMode: "HTML"})

	ctx.Data["songInfo"] = song

	return uploadToUser(b, ctx)
}

func externalUpload(b *gotgbot.Bot, ctx *ext.Context) error {
	fileLocation, ok := ctx.Data["fileLocation"].(string)
	if !ok {
		e := errors.New("no fileLocation")
		log.WithError(e).Error("internal error")

		_, _ = b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrUndefined(e), ctx.EffectiveUser.LanguageCode),
			&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
		return e
	}
	delete(ctx.Data, "fileLocation")

	f, err := os.Open(fileLocation)
	if err != nil {
		log.WithError(err).Error("can't open file to upload")
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	link, err := storage.Upload(f)
	if err != nil {
		log.WithError(err).Error("can't upload file to external storage")
		return err
	}

	_, err = b.SendMessage(
		ctx.EffectiveChat.Id,
		resp.Get(resp.ProcessStorageReady(link), ctx.EffectiveUser.LanguageCode),
		&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId},
	)
	if err != nil {
		log.WithError(err).Error("can't send message to user")
	}
	return nil
}

func uploadToUser(b *gotgbot.Bot, ctx *ext.Context) error {
	fileLocation, ok := ctx.Data["fileLocation"].(string)
	if !ok {
		e := errors.New("no fileLocation")
		log.WithError(e).Error("internal error")

		_, _ = b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrUndefined(e), ctx.EffectiveUser.LanguageCode),
			&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
		return e
	}
	delete(ctx.Data, "fileLocation")

	songInfo, ok := ctx.Data["songInfo"].(*soundcloader.Song)
	if !ok {
		e := errors.New("no songInfo")
		log.WithError(e).Error("internal error")

		_, _ = b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrUndefined(e), ctx.EffectiveUser.LanguageCode),
			&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
		return e
	}
	delete(ctx.Data, "songInfo")

	localLog := log.WithField("messageID", ctx.EffectiveMessage.MessageId)

	localLog.Info("fetched song, uploading to user..")

	f, e := os.Open(fileLocation)
	if e != nil {
		localLog.WithError(e).Error("can't open song file")

		_, _ = b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrUndefined(e), ctx.EffectiveUser.LanguageCode),
			&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
		return e
	}
	defer func() {
		_ = f.Close()
	}()

	_, err := b.SendAudio(ctx.EffectiveChat.Id,
		f,
		&gotgbot.SendAudioOpts{
			Caption:          "@scdl_info",
			Duration:         int64(songInfo.Duration.Seconds()),
			Performer:        songInfo.Author,
			Title:            songInfo.Title,
			Thumb:            songInfo.Thumbnail,
			ReplyToMessageId: ctx.EffectiveMessage.MessageId,
		},
	)

	return err
}
