package repositories

import (
	"context"
	"dtf/game_draw/internal/domain/models"
)

type SessionRepository interface {
	Save(ctx context.Context, session models.UserSession) error
	GetByEmail(ctx context.Context, email string) (models.UserSession, error)
}
