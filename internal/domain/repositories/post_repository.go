package repositories

import (
	"context"
	"dtf/game_draw/internal/domain/models"
	"time"
)

type PostRepository interface {
	SearchPosts(ctx context.Context, query string, dateFrom time.Time) ([]models.Post, error)
	ReactToPost(ctx context.Context, user models.DtfUserSession, post models.Post) error
	PostComment(ctx context.Context, user models.DtfUserSession, post models.Post, text string) error
}
