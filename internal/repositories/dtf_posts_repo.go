package repositories

import (
	"dtf/game_draw/internal/domain/models"
	"dtf/game_draw/pkg/dtfapi"
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

func (r dtfPostRepository) SearchPosts(query string, dateFrom time.Time) ([]models.Post, error) {
	news, err := r.dtfService.SearchNews(query, dateFrom)
	if err != nil {
		return nil, err
	}

	posts := make([]models.Post, 0, len(news))
	for _, newsItem := range news {
		post, err := models.FromDtfPost(newsItem)
		if err != nil {
			// TODO: add logger here
			continue
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (r dtfPostRepository) ReactToPost(post models.Post) error {
	err := r.dtfService.ReactToPost(int(post.Id))
	if err != nil {
		return err
	}

	return nil
}

func (r dtfPostRepository) PostComment(post models.Post, text string) error {
	err := r.dtfService.PostComment(int(post.Id), text)
	if err != nil {
		return err
	}

	return nil
}
