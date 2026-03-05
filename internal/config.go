package internal

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DbPath         string
	TelegramToken  string
	TelegramAdmins []int64
}

const configPath = ".env"

func NewConfig() (*Config, error) {
	env, err := godotenv.Read(configPath)
	if err != nil {
		return nil, err
	}

	telegramAdmins, err := envToTelegramAdmins(env["TELEGRAM_ADMINS"])
	if err != nil {
		return nil, err
	}

	sqlitePath := env["GOOSE_DBSTRING"]
	telegramToken := env["TELEGRAM_TOKEN"]

	config := &Config{
		DbPath:         sqlitePath,
		TelegramToken:  telegramToken,
		TelegramAdmins: telegramAdmins,
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

func envToTelegramAdmins(envStr string) ([]int64, error) {
	if envStr == "" {
		return make([]int64, 0), nil
	}

	strIds := strings.Split(envStr, ",")
	result := make([]int64, 0, len(strIds))
	errs := make([]error, 0)

	for _, v := range strIds {
		trimmedV := strings.TrimSpace(v)
		id, err := strconv.Atoi(trimmedV)
		if err != nil {
			errs = append(
				errs,
				fmt.Errorf("invalid admin telegram id=%v, err=%w", v, err),
			)
			continue
		}
		result = append(result, int64(id))
	}

	if len(errs) != 0 {
		return []int64{}, errors.Join(errs...)
	}

	return result, nil
}
