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

func ProcessURL(sender *mux.Sender, ctx *ext.Context) (success bool) {
	localLog := sender.Logger.With(zap.Int64("messageID", ctx.EffectiveMessage.MessageId))
	localLog.Info("checking url..")

	if len(ctx.EffectiveMessage.Entities) < 1 {
		// no url, shouldn't ever happen lol
		localLog.Info("no url here, finishing..")
		_, _ = sender.Reply(ctx, tr.ErrNotURL, true)
		return
	}

	receivedURL := getValidURL(ctx.EffectiveMessage)
	if receivedURL == "" {
		// non-valid url
		localLog.Info("no valid url here, finishing..")
		_, _ = sender.Reply(ctx, tr.ErrNotURL, true)
		return
	}

	tempMessage, err := sender.Reply(ctx, tr.ProcessStart, false)
	if err != nil {
		// can't send message after retrying, should stop now
		localLog.Info("can't send message to user", zap.Error(err))
		return
	}

	defer func() {
		_, _ = sender.DeleteMessage(context.Background(), tempMessage)
	}()

	urlInfo := soundcloader.Parse(receivedURL)
	if urlInfo == nil {
		// not SoundCloud url
		localLog.Info("no valid SoundCloud url here, finishing..")
		_, _ = sender.Reply(ctx, tr.ErrNotSC, true)
		return
	}

	return fetchSong(sender, ctx, urlInfo, tempMessage)
}

func fetchSong(sender *mux.Sender, ctx *ext.Context,
	urlInfo *soundcloader.URLInfo, tempMessage *gotgbot.Message,
) (success bool) {
	localLog := sender.Logger.With(
		zap.Int64("messageID", ctx.EffectiveMessage.MessageId),
		zap.String("url", urlInfo.String()),
	)

	var err error

	// tell user we start fetching song
	tempMessage, err = sender.EditMessage(sender.AppCtx, tempMessage,
		sender.Resp.Get(tr.ProcessFetching, tr.GetLang(ctx.EffectiveMessage)), nil)
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
			_, _ = sender.Reply(ctx, tr.ErrUnsupportedFormat, false)
			return
		}

		localLog.Error("undefined error while fetching song info, exiting..", zap.Error(err))
		_, _ = sender.Reply(ctx, tr.ErrUndefined(err), false)
		return
	}

	location, err := song.GetNext()
	if err != nil {
		if err == soundcloader.EmptyStream {
			localLog.Info("song unavailable, exiting..")
			_, _ = sender.Reply(ctx, tr.ErrUnavailableSong, false)
			return
		}

		localLog.Error("undefined error while downloading song, exiting..", zap.Error(err))
		_, _ = sender.Reply(ctx, tr.ErrUndefined(err), false)
		return
	}

	songStats, err := os.Stat(location)
	if err != nil {
		// shouldn't ever happen, but means file isn't available, so we can't continue
		localLog.Info("can't access file, exiting..")
		_, _ = sender.Reply(ctx, tr.ErrInternal(fmt.Errorf("can't read file")), false)
		return
	}

	if songStats.Size() > MaxSize {
		// oversize, can't send via telegram
		localLog.Info("size limit, uploading to external storage..")
		_, _ = sender.EditMessage(sender.AppCtx, tempMessage,
			sender.Resp.Get(tr.ProcessStorage, tr.GetLang(ctx.EffectiveMessage)), nil)

		link, err := sender.External.Upload(location)
		if err != nil {
			_, _ = sender.Reply(ctx, tr.ErrInternal(err), false)
			return
		}

		_, _ = sender.Reply(ctx, tr.ProcessStorageReady(link), false)
		return true
	}

	tempMessage, err = sender.EditMessage(context.Background(), tempMessage,
		sender.Resp.Get(tr.ProcessUploading, tr.GetLang(ctx.EffectiveMessage)), nil)
	if err != nil {
		// can't edit message after retrying, should stop now
		localLog.Info("can't edit temp message", zap.Error(err))
		return
	}

	return uploadSong(sender, ctx, song, location)
}

func uploadSong(sender *mux.Sender, ctx *ext.Context, songInfo *soundcloader.Song, location string) (success bool) {
	localLog := sender.Logger.With(zap.Int64("messageID", ctx.EffectiveMessage.MessageId))
	localLog.Info("fetched song, uploading to user..")

	f, e := os.Open(location)
	if e != nil {
		localLog.Error("can't open song file", zap.Error(e))

		_, _ = sender.Reply(ctx, tr.ErrInternal(fmt.Errorf("can't open file")), false)
		return
	}
	defer func() {
		_ = f.Close()
	}()

	songCaption := sender.Resp.Get(tr.ProcessReady(
		songInfo.Thumbnail,
		songInfo.Permalink,
		//fmt.Sprintf("https://youtu.be/%s", video.ID),
	), tr.GetLang(ctx.EffectiveMessage))

	if _, err := sender.SendAudio(sender.AppCtx, ctx.EffectiveChat.Id, f,
		&gotgbot.SendAudioOpts{
			ReplyToMessageId: ctx.EffectiveMessage.MessageId,
			Caption:          songCaption,
			Title:            songInfo.Title,
			Performer:        songInfo.Author,
			Duration:         int64(songInfo.Duration.Seconds()),
			Thumb:            songInfo.Thumbnail,
		},
	); err != nil {
		localLog.Error("can't send song to user", zap.Error(e))
	}

	return true
}
