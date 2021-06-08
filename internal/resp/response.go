package tr

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml"
	"golang.org/x/text/language"
	"path/filepath"
)

type Translation struct {
	bundle *i18n.Bundle
}

func New(homePath string) *Translation {
	b := i18n.NewBundle(language.English)
	b.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	b.MustLoadMessageFile(filepath.Join(homePath, "en.toml"))
	b.MustLoadMessageFile(filepath.Join(homePath, "ru.toml"))

	return &Translation{
		bundle: b,
	}
}

func (t *Translation) Get(s *i18n.LocalizeConfig, lang string) string {
	return i18n.NewLocalizer(t.bundle, lang).MustLocalize(s)
}

func GetLang(msg *gotgbot.Message) string {
	if msg.From == nil {
		return ""
	}
	return msg.From.LanguageCode
}

// Process phrases
var (
	ProcessStart        = &i18n.LocalizeConfig{MessageID: "process.Start"}
	ProcessFetching     = &i18n.LocalizeConfig{MessageID: "process.Fetching"}
	ProcessUploading    = &i18n.LocalizeConfig{MessageID: "process.Uploading"}
	ProcessStorage      = &i18n.LocalizeConfig{MessageID: "process.Storage"}
	ProcessStorageReady = func(s string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			MessageID:    "process.StorageReady",
			TemplateData: map[string]string{"link": s},
		}
	}
)

// Command-related phrases
var (
	CmdStart     = &i18n.LocalizeConfig{MessageID: "cmd.Start"}
	CmdHelp      = &i18n.LocalizeConfig{MessageID: "cmd.Help"}
	CmdUndefined = &i18n.LocalizeConfig{MessageID: "cmd.Default"}
)

// Error messages
var (
	ErrNotURL            = &i18n.LocalizeConfig{MessageID: "err.NotURL"}
	ErrNotSCURL          = &i18n.LocalizeConfig{MessageID: "err.NotSoundCloudURL"}
	ErrUnsupportedFormat = &i18n.LocalizeConfig{MessageID: "err.UnsupportedFormat"}
	ErrUnavailableSong   = &i18n.LocalizeConfig{MessageID: "err.UnavailableSong"}
	ErrInternal          = func(err error) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			MessageID:    "err.Internal",
			TemplateData: map[string]string{"errMessage": err.Error()},
		}
	}
	ErrUndefined = func(err error) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			MessageID:    "err.Default",
			TemplateData: map[string]string{"errMessage": err.Error()},
		}
	}
)

// Util text
var (
	UtilGetCover = &i18n.LocalizeConfig{MessageID: "util.GetCover"}
)
