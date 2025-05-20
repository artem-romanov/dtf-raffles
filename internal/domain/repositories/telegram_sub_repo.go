package repositories

import (
	"context"
	"dtf/game_draw/internal/domain"
	"dtf/game_draw/internal/domain/models"
)

type TelegramSubscribersRepository interface {
	// selectors
	FindById(ctx context.Context, telegramId int64) (models.TelegramSession, error)
	GetAll(ctx context.Context) ([]models.TelegramSession, error)

	// mutators
	RegisterUser(ctx context.Context, tx domain.DBTX, telegramId int64) error
	UnregisterUser(ctx context.Context, tx domain.DBTX, telegramId int64) error
}
