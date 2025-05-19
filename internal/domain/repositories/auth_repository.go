package repositories

import "dtf/game_draw/internal/domain/models"

type AuthRepository interface {
	Login(email, password string) (models.UserSession, error)
	RefreshToken(user models.UserSession) (models.UserSession, error)
}
