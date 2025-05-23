package usecases

import (
	"context"
	"dtf/game_draw/internal/domain/models"
	"dtf/game_draw/internal/domain/repositories"
	"strings"
	"time"
)

type GetActiveRafflePostsUseCase struct {
	postRepo repositories.PostRepository
}

func NewGetActiveRafflePostsUseCase(repo repositories.PostRepository) *GetActiveRafflePostsUseCase {
	return &GetActiveRafflePostsUseCase{
		postRepo: repo,
	}
}

func (uc *GetActiveRafflePostsUseCase) Execute(ctx context.Context, fromDate time.Time) ([]models.Post, error) {
	// getting all raffle (Розыгрыш) posts
	posts, err := uc.postRepo.SearchPosts(ctx, "Розыгрыш", fromDate)
	if err != nil {
		return nil, err
	}

	// filtering and keeping only ongoing posts
	var result []models.Post
	for _, post := range posts {
		if isEnded(post.Title) || post.RepliedTo.Valid {
			continue
		}
		result = append(result, post)
	}

	return result, nil
}

var EndedRaffleKeywords = []string{
	"завершен",
	"завершён",
	"закончен",
	"закончился",
	"итоги",
	"подведены итоги",
	"победители",
	"разыграли",
	"финал",
	"результа",
	"завершили",
}

func isEnded(title string) bool {
	title = strings.ToLower(title)
	for _, keyWord := range EndedRaffleKeywords {
		if strings.Contains(title, keyWord) {
			return true
		}
	}

	return false
}
