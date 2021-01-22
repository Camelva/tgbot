package main

import (
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

type Config struct {
	Telegram Telegram `yaml:"telegram"`
	Mode     string   `yaml:"mode" envconfig:"MODE"`
}

type Telegram struct {
	Token     string `yaml:"token" envconfig:"TELEGRAM_TOKEN"`
	TestToken string `yaml:"testToken" envconfig:"TELEGRAM_TEST_TOKEN"`
	OwnerID   int    `yaml:"ownerID" envconfig:"TELEGRAM_OWNER_ID"`
}

func loadConfig(file string) (cfg Config) {
	// env variables always overwrite
	readFile(file, &cfg)
	readEnv(&cfg)

	if (cfg == Config{}) {
		log.Fatal("config not loaded!")
	}
	if cfg.Mode == "debug" {
		cfg.Telegram.Token = cfg.Telegram.TestToken
	}
	return cfg
}

func readFile(fileName string, cfg *Config) {
	f, err := os.Open(fileName)
	if err != nil {
		log.Println("can't read config file")
	}
	defer closeFile(f)

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
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

func closeFile(f *os.File) {
	err := f.Close()

	if err != nil {
		log.Printf("error while closing file: %v\n", err)
	}
}
