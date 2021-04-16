package resp

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml"
	"golang.org/x/text/language"
)

var bundle i18n.Bundle

func init() {
	var b = i18n.NewBundle(language.English)
	b.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	b.MustLoadMessageFile("resp/en.toml")
	b.MustLoadMessageFile("resp/ru.toml")

	bundle = *b
}

func Get(s *i18n.LocalizeConfig, lang string) string {
	return i18n.NewLocalizer(&bundle, lang).MustLocalize(s)
}

var (
	ProcessStart     = &i18n.LocalizeConfig{MessageID: "process.Start"}
	ProcessQueue     = &i18n.LocalizeConfig{MessageID: "process.Queue"}
	ProcessFetching  = &i18n.LocalizeConfig{MessageID: "process.Fetching"}
	ProcessUploading = &i18n.LocalizeConfig{MessageID: "process.Uploading"}

	CmdStart     = &i18n.LocalizeConfig{MessageID: "cmd.Start"}
	CmdHelp      = &i18n.LocalizeConfig{MessageID: "cmd.Help"}
	CmdUndefined = &i18n.LocalizeConfig{MessageID: "cmd.Default"}
	//
	ErrNotURL            = &i18n.LocalizeConfig{MessageID: "err.NotURL"}
	ErrNotSCURL          = &i18n.LocalizeConfig{MessageID: "err.NotSoundCloudURL"}
	ErrUnsupportedFormat = &i18n.LocalizeConfig{MessageID: "err.UnsupportedFormat"}
	//errUnsupportedService = "err.UnsupportedService"
	ErrSizeLimit       = &i18n.LocalizeConfig{MessageID: "err.SizeLimit"}
	ErrUnavailableSong = &i18n.LocalizeConfig{MessageID: "err.UnavailableSong"}
	ErrUndefined       = func(err error) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			MessageID:    "err.Default",
			TemplateData: map[string]string{"errMessage": err.Error()},
		}
	}
)
