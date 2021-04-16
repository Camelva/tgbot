package main

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

type Config struct {
	Telegram struct {
		Token   string
		OwnerID int64
	}
	Storage struct {
		Token string
	}
	Mode string
}

func loadConfigs(fileName string) (cfg Config) {
	if err := godotenv.Load(fileName); err != nil {
		log.Info("can't load config file")
	}

	if mode := os.Getenv("TGBOT_MODE"); mode != "" {
		cfg.Mode = mode
	}

	if ownerID := os.Getenv("TGBOT_OWNER"); ownerID != "" {
		id, err := strconv.ParseInt(ownerID, 10, 64)
		if err == nil {
			cfg.Telegram.OwnerID = id
		}
	}

	if cfg.Mode == "debug" {
		if token := os.Getenv("TGBOT_TOKEN_TEST"); token != "" {
			cfg.Telegram.Token = token
		}
	} else {
		if token := os.Getenv("TGBOT_TOKEN"); token != "" {
			cfg.Telegram.Token = token
		}
	}

	if cfg.Telegram.Token == "" {
		log.Fatal("token is required")
	}

	if token := os.Getenv("TGBOT_STORAGE_TOKEN"); token != "" {
		cfg.Storage.Token = token
	}
	return cfg
}
