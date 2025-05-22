package internal

import (
	"errors"

	"github.com/joho/godotenv"
)

type Config struct {
	DbPath        string
	TelegramToken string
}

const configPath = ".env"

func NewConfig() (*Config, error) {
	env, err := godotenv.Read(configPath)
	if err != nil {
		return nil, err
	}

	sqlitePath := env["GOOSE_DBSTRING"]
	telegramToken := env["TELEGRAM_TOKEN"]

	config := &Config{
		DbPath:        sqlitePath,
		TelegramToken: telegramToken,
	}

	err = validateConfig(*config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func validateConfig(config Config) error {
	if config.DbPath == "" {
		return errors.New("Sqlite Db path is absent")
	}

	if config.TelegramToken == "" {
		return errors.New("Telegram token is absent")
	}

	return nil
}
