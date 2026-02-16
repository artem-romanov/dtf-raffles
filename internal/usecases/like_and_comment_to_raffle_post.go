package usecases

import (
	"context"
	"dtf/game_draw/internal/domain/managers"
	"dtf/game_draw/internal/domain/models"
	"dtf/game_draw/internal/domain/repositories"
)

type LikeAndPostToRafflePostUseCase struct {
	postRepo    repositories.PostRepository
	userManager managers.UserManager
}

func NewLikeAndPostToRafflePostUseCase(
	postRepo repositories.PostRepository,
	userManager managers.UserManager,
) *LikeAndPostToRafflePostUseCase {
	return &LikeAndPostToRafflePostUseCase{
		postRepo:    postRepo,
		userManager: userManager,
	}
}

func (uc *LikeAndPostToRafflePostUseCase) Execute(ctx context.Context, userEmail string, post models.Post) error {
	user, err := uc.userManager.BuildSession(ctx, userEmail)
	if err != nil {
		return err
	}

	err = uc.postRepo.ReactToPost(ctx, user, post)
	if err != nil {
		return err
	}

	err = uc.postRepo.PostComment(ctx, user, post, "Участвую")
	if err != nil {
		return err
	}

	return nil
}
