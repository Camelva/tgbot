package scloud

import (
	"context"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/camelva/soundcloader"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strconv"
	"tgbot/internal/mux"
	"tgbot/internal/resp"
)

var MaxSize int64 = 50 * 1024 * 1024

func ProcessURL(sender *mux.Sender, ctx *ext.Context) {
	localLog := sender.Logger.With(zap.Int64("messageID", ctx.EffectiveMessage.MessageId))
	localLog.Info("checking url..")

	tempMessage, err := sender.ReplyToMessage(
		context.Background(),
		ctx.EffectiveMessage,
		sender.Resp.Get(tr.ProcessStart, tr.GetLang(ctx.Message)),
		nil,
	)
	if err != nil {
		// can't send message after retrying, should stop now
		localLog.Info("can't send message to user", zap.Error(err))
		return
	}

	defer func() {
		_, _ = sender.DeleteMessage(context.Background(), tempMessage)
	}()

	if len(ctx.EffectiveMessage.Entities) < 1 {
		// no url, shouldn't ever happen lol
		localLog.Info("no url here, finishing..")
		_, _ = sender.ReplyToMessage(
			context.Background(),
			ctx.EffectiveMessage,
			sender.Resp.Get(tr.ErrNotURL, tr.GetLang(ctx.Message)),
			nil,
		)
		return
	}

	receivedURL := getValidURL(ctx.EffectiveMessage)
	if receivedURL == "" {
		// non-valid url
		localLog.Info("no valid url here, finishing..")
		_, _ = sender.ReplyToMessage(
			context.Background(),
			ctx.EffectiveMessage,
			sender.Resp.Get(tr.ErrNotURL, tr.GetLang(ctx.Message)),
			nil,
		)
		return
	}

	urlInfo := soundcloader.Parse(receivedURL)
	if urlInfo == nil {
		// not SoundCloud url
		localLog.Info("no valid SoundCloud url here, finishing..")
		_, _ = sender.ReplyToMessage(
			context.Background(),
			ctx.EffectiveMessage,
			sender.Resp.Get(tr.ErrNotSCURL, tr.GetLang(ctx.Message)),
			nil,
		)
		return
	}

	fetchSong(sender, ctx, urlInfo, tempMessage)
	return
}

func fetchSong(sender *mux.Sender, ctx *ext.Context, urlInfo *soundcloader.URLInfo, tempMessage *gotgbot.Message) {
	localLog := sender.Logger.With(
		zap.Int64("messageID", ctx.EffectiveMessage.MessageId),
		zap.String("url", urlInfo.String()),
	)

	var err error

	// tell user we start fetching song
	tempMessage, err = sender.EditMessage(
		context.Background(),
		tempMessage,
		sender.Resp.Get(tr.ProcessFetching, tr.GetLang(ctx.Message)),
		nil,
	)
	if err != nil {
		// can't send message after retrying, should stop now
		localLog.Info("can't send message to user", zap.Error(err))
		return
	}

	// making separate folder for each message to avoid conflicts
	sClient := *soundcloader.DefaultClient
	sClient.OutputFolder = filepath.Join(sClient.OutputFolder, strconv.FormatInt(ctx.UpdateId, 10))

	defer func() {
		_ = os.RemoveAll(sClient.OutputFolder)
	}()

	song, err := sClient.GetURL(urlInfo)
	if err != nil {
		if err == soundcloader.NotSong {
			localLog.Info("not song url, exiting..")
			_, _ = sender.ReplyToMessage(
				context.Background(),
				ctx.EffectiveMessage,
				sender.Resp.Get(tr.ErrUnsupportedFormat, tr.GetLang(ctx.Message)),
				nil,
			)
			return
		}
		localLog.Error("undefined error while fetching song info, exiting..", zap.Error(err))
		_, _ = sender.ReplyToMessage(
			context.Background(),
			ctx.EffectiveMessage,
			sender.Resp.Get(tr.ErrUndefined(err), tr.GetLang(ctx.Message)),
			nil,
		)
		return
	}

	location, err := song.GetNext()
	if err != nil {
		if err == soundcloader.EmptyStream {
			localLog.Info("song unavailable, exiting..")
			_, _ = sender.ReplyToMessage(
				context.Background(),
				ctx.EffectiveMessage,
				sender.Resp.Get(tr.ErrUnavailableSong, tr.GetLang(ctx.Message)),
				nil,
			)
			return
		}
		localLog.Error("undefined error while downloading song, exiting..", zap.Error(err))
		_, _ = sender.ReplyToMessage(
			context.Background(),
			ctx.EffectiveMessage,
			sender.Resp.Get(tr.ErrUndefined(err), tr.GetLang(ctx.Message)),
			nil,
		)
		return
	}

	songStats, err := os.Stat(location)
	if err != nil {
		// shouldn't ever happen, but means file isn't available, can't continue
		localLog.Info("can't access file, exiting..")
		_, _ = sender.ReplyToMessage(
			context.Background(),
			ctx.EffectiveMessage,
			sender.Resp.Get(tr.ErrInternal(fmt.Errorf("can't read file")), tr.GetLang(ctx.Message)),
			nil,
		)
		return
	}

	if songStats.Size() > MaxSize {
		// oversize, can't send via telegram
		localLog.Info("size limit, uploading to external storage..")
		_, _ = sender.EditMessage(
			context.Background(),
			tempMessage,
			sender.Resp.Get(tr.ProcessStorage, tr.GetLang(ctx.Message)),
			nil,
		)

		link, err := sender.External.Upload(location)
		if err != nil {
			_, _ = sender.ReplyToMessage(
				context.Background(),
				ctx.EffectiveMessage,
				sender.Resp.Get(tr.ErrInternal(err), tr.GetLang(ctx.Message)),
				nil,
			)
		} else {
			_, _ = sender.ReplyToMessage(
				context.Background(),
				ctx.EffectiveMessage,
				sender.Resp.Get(tr.ProcessStorageReady(link), tr.GetLang(ctx.Message)),
				nil,
			)
		}

		return
	}

	if _, err = sender.EditMessage(
		context.Background(),
		tempMessage,
		sender.Resp.Get(tr.ProcessUploading, tr.GetLang(ctx.Message)),
		nil,
	); err != nil {
		// can't edit message after retrying, should stop now
		localLog.Info("can't edit temp message", zap.Error(err))
		return
	}

	uploadSong(sender, ctx, song, location)
	return
}

func uploadSong(sender *mux.Sender, ctx *ext.Context, songInfo *soundcloader.Song, location string) {
	localLog := sender.Logger.With(zap.Int64("messageID", ctx.EffectiveMessage.MessageId))
	localLog.Info("fetched song, uploading to user..")

	f, e := os.Open(location)
	if e != nil {
		localLog.Error("can't open song file", zap.Error(e))

		_, _ = sender.ReplyToMessage(
			context.Background(),
			ctx.EffectiveMessage,
			sender.Resp.Get(tr.ErrInternal(fmt.Errorf("can't open file")), tr.GetLang(ctx.Message)),
			nil,
		)
		return
	}
	defer func() {
		_ = f.Close()
	}()

	songCaption := fmt.Sprintf(
		"<a href=\"%s\">%s</a>\n\n@scdl_info",
		songInfo.Thumbnail,
		sender.Resp.Get(tr.UtilGetCover, tr.GetLang(ctx.Message)),
	)

	if _, err := sender.SendAudio(
		context.Background(),
		ctx.EffectiveMessage.Chat.Id,
		f,
		&gotgbot.SendAudioOpts{
			Caption:          songCaption,
			ParseMode:        "HTML",
			Duration:         int64(songInfo.Duration.Seconds()),
			Performer:        songInfo.Author,
			Title:            songInfo.Title,
			Thumb:            songInfo.Thumbnail,
			ReplyToMessageId: ctx.EffectiveMessage.MessageId,
		},
	); err != nil {
		localLog.Error("can't send song to user", zap.Error(e))
	}

	return
}
