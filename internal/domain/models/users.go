package models

import "time"

type DtfUserSession struct {
	Email            string
	AccessToken      string
	RefreshToken     string
	AccessExpiration time.Time
}

type DtfUserInfo struct {
	Id   int
	Name string
	Url  string
}

type TelegramSession struct {
	TelegramId int64
	CreatedAt  time.Time
}
