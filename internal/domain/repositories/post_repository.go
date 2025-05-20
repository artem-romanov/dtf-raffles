package repositories

import (
	"dtf/game_draw/internal/domain/models"
	"time"
)

// TODO: add context to all methods
type PostRepository interface {
	SearchPosts(query string, dateFrom time.Time) ([]models.Post, error)
	ReactToPost(user models.DtfUserSession, post models.Post) error
	PostComment(user models.DtfUserSession, post models.Post, text string) error
}
