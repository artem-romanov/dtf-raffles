package usecases

import (
	"dtf/game_draw/internal/domain/models"
	"dtf/game_draw/internal/domain/repositories"
)

type LikeAndPostToRafflePostUseCase struct {
	postRepo repositories.PostRepository
}

func NewLikeAndPostToRafflePostUseCase() LikeAndPostToRafflePostUseCase {
	return LikeAndPostToRafflePostUseCase{}
}

func (uc *LikeAndPostToRafflePostUseCase) Execute(post models.Post) error {
	err := uc.postRepo.ReactToPost(post)
	if err != nil {
		return err
	}
	err = uc.postRepo.PostComment(post, "Участвую")
	if err != nil {
		return err
	}
	return nil
}
