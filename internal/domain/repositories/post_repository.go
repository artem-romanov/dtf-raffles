package repositories

import (
	"dtf/game_draw/internal/domain/models"
	"time"
)

type PostRepository interface {
	SearchPosts(query string, dateFrom time.Time) ([]models.Post, error)
	ReactToPost(post models.Post, user models.UserSession) (models.UserSession, error)
	PostComment(post models.Post, text string, user models.UserSession) (models.UserSession, error)
}
