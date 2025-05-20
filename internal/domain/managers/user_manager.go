package managers

import (
	"context"
	"dtf/game_draw/internal/domain/models"
)

type UserManager interface {
	BuildSession(ctx context.Context, email string) (models.DtfUserSession, error)
	EmailLogin(ctx context.Context, email, password string) (models.DtfUserSession, error)
}
