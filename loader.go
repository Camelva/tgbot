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
)

func loadSong(b *gotgbot.Bot, ctx *ext.Context) (err error) {
	tempMessage, ok := ctx.Data["tempMessage"].(*gotgbot.Message)
	if !ok {
		e := errors.New("no tempMessage")
		_, _ = b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrUndefined(e), ctx.EffectiveUser.LanguageCode),
			&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
		return e
	}
	delete(ctx.Data, "tempMessage")

	parsedURL, ok := ctx.Data["parsedURL"].(*soundcloader.URLInfo)
	if !ok {
		e := errors.New("no parsedURL")
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

	song, e := sClient.GetURL(parsedURL)
	if e != nil {
		if e == soundcloader.NotSong {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrUnsupportedFormat, ctx.EffectiveUser.LanguageCode),
				&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
		} else {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrUndefined(e), ctx.EffectiveUser.LanguageCode),
				&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
		}
		return err
	}

	location, err := song.GetNext()
	if err != nil {
		if err == soundcloader.EmptyStream {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrUnavailableSong, ctx.EffectiveUser.LanguageCode),
				&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId})
		} else {
			_, _ = b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrUndefined(err), ctx.EffectiveUser.LanguageCode),
				&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId})
		}
		return err
	}
	tempMessage, err = tempMessage.EditText(b, resp.Get(resp.ProcessUploading, ctx.EffectiveUser.LanguageCode),
		&gotgbot.EditMessageTextOpts{ParseMode: "HTML"})

	ctx.Data["fileLocation"] = location
	ctx.Data["songInfo"] = song

	log.WithField("message", ctx.EffectiveMessage.Text).Info("done, uploading song to user..")

	return uploadToUser(b, ctx)
}

func uploadToUser(b *gotgbot.Bot, ctx *ext.Context) error {
	fileLocation, ok := ctx.Data["fileLocation"].(string)
	if !ok {
		e := errors.New("no fileLocation")
		_, _ = b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrUndefined(e), ctx.EffectiveUser.LanguageCode),
			&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
		return e
	}
	delete(ctx.Data, "fileLocation")

	songInfo, ok := ctx.Data["songInfo"].(*soundcloader.Song)
	if !ok {
		e := errors.New("no songInfo")
		_, _ = b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrUndefined(e), ctx.EffectiveUser.LanguageCode),
			&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
		return e
	}
	delete(ctx.Data, "songInfo")

	f, e := os.Open(fileLocation)
	defer func() {
		_ = f.Close()
		_ = os.Remove(fileLocation)
	}()
	if e != nil {
		_, _ = b.SendMessage(ctx.EffectiveChat.Id, resp.Get(resp.ErrUndefined(e), ctx.EffectiveUser.LanguageCode),
			&gotgbot.SendMessageOpts{ReplyToMessageId: ctx.EffectiveMessage.MessageId, ParseMode: "HTML"})
		return e
	}

	_, err := b.SendAudio(ctx.EffectiveChat.Id,
		gotgbot.NamedFile{
			File:     f,
			FileName: songInfo.Permalink,
		},
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
