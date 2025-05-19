package models

import "time"

type UserSession struct {
	Email            string
	AccessToken      string
	RefreshToken     string
	AccessExpiration time.Time
}
