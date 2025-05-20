package repositories

import (
	"context"
	"dtf/game_draw/internal/domain"
	"dtf/game_draw/internal/domain/models"
)

type DtfSessionRepository interface {
	// getters
	GetByEmail(ctx context.Context, email string) (models.DtfUserSession, error)

	// mutators
	Save(ctx context.Context, tx domain.DBTX, session models.DtfUserSession) error
	DeleteByEmail(ctx context.Context, tx domain.DBTX, email string) error
}
