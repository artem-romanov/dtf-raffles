package repositories

import (
	"context"
	"dtf/game_draw/internal/domain/models"
)

type AuthRepository interface {
	Login(ctx context.Context, email, password string) (models.DtfUserSession, error)
	RefreshToken(ctx context.Context, user models.DtfUserSession) (models.DtfUserSession, error)
	SelfInfo(ctx context.Context, user models.DtfUserSession) (models.DtfUserInfo, error)
}
