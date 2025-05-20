package internal

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DbPath        string
	TelegramToken string
}

func NewConfig() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, err
	}

	// TODO: add validation checks for each required field
	return &Config{
		DbPath:        os.Getenv("GOOSE_DBSTRING"),
		TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
	}, nil
}
