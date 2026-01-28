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
	errs := []error{}

	if config.DbPath == "" {
		errs = append(errs, errors.New("sqlite Db path is missing"))
	}

	if config.TelegramToken == "" {
		errs = append(errs, errors.New("telegram token is missing"))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
