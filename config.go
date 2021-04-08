package main

import (
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Telegram Telegram `yaml:"telegram"`
	Mode     string   `yaml:"mode" envconfig:"MODE"`

	Settings Settings
}

type Settings struct {
	LoadersLimit int    `yaml:"loadersLimit"`
	FileTTL      string `yaml:"fileTTL"`
}

type Telegram struct {
	Token     string `yaml:"token" envconfig:"TELEGRAM_TOKEN"`
	TestToken string `yaml:"testToken" envconfig:"TELEGRAM_TEST_TOKEN"`
	OwnerID   int    `yaml:"ownerID" envconfig:"TELEGRAM_OWNER_ID"`
}

func loadConfigs(envFile, configFile string) (cfg Config) {
	// env variables always overwrite
	readConfigFromFile("env", envFile, &cfg)
	readEnv(&cfg)
	readConfigFromFile("settings", configFile, &cfg)

	if (cfg == Config{}) {
		log.Fatal("config not loaded!")
	}
	if cfg.Mode == "debug" {
		cfg.Telegram.Token = cfg.Telegram.TestToken
	}
	return cfg
}

func readEnv(cfg *Config) {
	err := envconfig.Process("", cfg)
	if err != nil {
		log.WithError(err).Warn("can't read environment variables")
	}
	return
}

func readConfigFromFile(mode string, configFile string, cfg *Config) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.WithError(err).Warn("can't read config file")
	}

	if mode == "env" {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			log.WithError(err).Warn("can't decode config file")
		}
	} else if mode == "settings" {
		conf := new(Settings)

		if err := yaml.Unmarshal(data, conf); err != nil {
			log.WithError(err).Warn("can't decode config file")
		}
		cfg.Settings = *conf
	}
	return
}

func changeConfig(s Settings) error {
	var conf = new(Settings)
	data, err := ioutil.ReadFile("config.yml")
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, conf); err != nil {
		return err
	}

	if s.FileTTL != "" {
		conf.FileTTL = s.FileTTL
	}
	if s.LoadersLimit != 0 {
		conf.LoadersLimit = s.LoadersLimit
	}

	res, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile("config.yml", res, 0755); err != nil {
		return err
	}
	return nil
}
