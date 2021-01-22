package main

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	//"github.com/BurntSushi/toml"
	"github.com/pelletier/go-toml"
	"golang.org/x/text/language"
)

var (
	processStart     = &i18n.LocalizeConfig{MessageID: "process.Start"}
	processFetching  = &i18n.LocalizeConfig{MessageID: "process.Fetching"}
	processUploading = &i18n.LocalizeConfig{MessageID: "process.Uploading"}

	cmdStart     = &i18n.LocalizeConfig{MessageID: "cmd.Start"}
	cmdHelp      = &i18n.LocalizeConfig{MessageID: "cmd.Help"}
	cmdUndefined = &i18n.LocalizeConfig{MessageID: "cmd.Default"}

	errNotURL            = &i18n.LocalizeConfig{MessageID: "err.NotURL"}
	errNotSCURL          = &i18n.LocalizeConfig{MessageID: "err.NotSoundCloudURL"}
	errUnsupportedFormat = &i18n.LocalizeConfig{MessageID: "err.UnsupportedFormat"}
	//errUnsupportedService = "err.UnsupportedService"
	errSizeLimit       = &i18n.LocalizeConfig{MessageID: "err.SizeLimit"}
	errUnavailableSong = &i18n.LocalizeConfig{MessageID: "err.UnavailableSong"}
	errUndefined       = func(err error) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			MessageID:    "err.Default",
			TemplateData: map[string]string{"errMessage": err.Error()},
		}
	}
)

func getResponses() *i18n.Bundle {
	var bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile("en.toml")
	bundle.MustLoadMessageFile("ru.toml")

	return bundle
}
