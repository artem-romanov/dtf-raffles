package repositories

import (
	"dtf/game_draw/internal/domain/models"
	"time"
)

type PostRepository interface {
	SearchPosts(query string, dateFrom time.Time) ([]models.Post, error)
	ReactToPost(user models.UserSession, post models.Post) error
	PostComment(user models.UserSession, post models.Post, text string) error
}
