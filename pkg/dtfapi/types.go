package dtfapi

import (
	"time"
)

type Tokens struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiration time.Time
}

func (t Tokens) IsAccessValid() bool {
	if t.AccessToken == "" {
		return false
	}

	diff := time.Until(t.AccessExpiration)
	if diff.Microseconds() <= 0 {
		return false
	}

	return true
}

type UserSession struct {
	Email      string
	UserTokens Tokens
}

type BlogPost struct {
	Id    int
	Title string
	Uri   string
}
