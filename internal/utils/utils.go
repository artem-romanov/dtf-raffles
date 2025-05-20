package utils

import (
	"dtf/game_draw/internal/domain/models"
	"time"
)

func UserExpired(session models.DtfUserSession) bool {
	if session.AccessToken == "" {
		return true
	}

	diff := time.Until(session.AccessExpiration)
	if diff.Microseconds() <= 0 {
		return true
	}

	return false
}
