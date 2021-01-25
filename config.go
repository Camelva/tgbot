package main

import (
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type Config struct {
	Telegram Telegram `yaml:"telegram"`
	Mode     string   `yaml:"mode" envconfig:"MODE"`

	Settings Settings `yaml:"settings"`
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
	readFile(envFile, &cfg)
	readEnv(&cfg)
	readFile(configFile, &cfg)

	if (cfg == Config{}) {
		log.Fatal("config not loaded!")
	}
	if cfg.Mode == "debug" {
		cfg.Telegram.Token = cfg.Telegram.TestToken
	}
	return cfg
}

func readFile(fileName string, cfg *Config) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Println("can't read config file")
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		log.Println("can't decode config file")
	}
	return
}

func readEnv(cfg *Config) {
	err := envconfig.Process("", cfg)
	if err != nil {
		log.Println("can't read environment variables")
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
