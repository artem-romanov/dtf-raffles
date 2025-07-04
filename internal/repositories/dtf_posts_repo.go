package repositories

import (
	"context"
	"dtf/game_draw/internal/domain/models"
	"dtf/game_draw/pkg/dtfapi"
	"log/slog"
	"time"
)

type dtfPostRepository struct {
	dtfService *dtfapi.DtfService
}

func NewDtfPostRepository(dtfService *dtfapi.DtfService) *dtfPostRepository {
	return &dtfPostRepository{
		dtfService: dtfService,
	}
}

func (r dtfPostRepository) SearchPosts(ctx context.Context, query string, dateFrom time.Time) ([]models.Post, error) {
	news, err := r.dtfService.SearchNews(ctx, query, dateFrom)
	if err != nil {
		return nil, err
	}

	posts := make([]models.Post, 0, len(news))
	for _, newsItem := range news {
		post, err := models.FromDtfPost(newsItem)
		if err != nil {
			slog.Warn("Post can't be parsed.", "post", newsItem)
			continue
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (r dtfPostRepository) ReactToPost(ctx context.Context, user models.DtfUserSession, post models.Post) error {
	err := r.dtfService.ReactToPost(ctx, user.AccessToken, int(post.Id))
	if err != nil {
		return err
	}

	return nil
}

func (r dtfPostRepository) PostComment(ctx context.Context, user models.DtfUserSession, post models.Post, text string) error {
	err := r.dtfService.PostComment(ctx, user.AccessToken, int(post.Id), text)
	if err != nil {
		return err
	}

	return nil
}
