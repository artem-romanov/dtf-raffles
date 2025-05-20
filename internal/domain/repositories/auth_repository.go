package repositories

import "dtf/game_draw/internal/domain/models"

// TODO: add context to all methods
type AuthRepository interface {
	Login(email, password string) (models.DtfUserSession, error)
	RefreshToken(user models.DtfUserSession) (models.DtfUserSession, error)
	SelfInfo(user models.DtfUserSession) (models.DtfUserInfo, error)
}
