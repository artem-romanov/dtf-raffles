package main

import (
	"context"
	"dtf/game_draw/internal/domain/managers"
	"dtf/game_draw/internal/domain/models"
	"dtf/game_draw/internal/domain/repositories"
	implManagers "dtf/game_draw/internal/managers"
	implRepo "dtf/game_draw/internal/repositories"
	"dtf/game_draw/internal/storage/sqlite"
	"dtf/game_draw/internal/usecases"
	"dtf/game_draw/pkg/dtfapi"
	"fmt"
	"log"
	"time"
)

const (
	email    string = "xenox2048@gmail.com"
	password string = "Azxsdcvf12"
)

func main() {
	ctx := context.Background()

	dtfClient := dtfapi.NewClient(ctx)
	dtfService := dtfapi.NewService(dtfClient.Client())

	db, err := sqlite.InitDB("./dtf_db.sqlite")
	if err != nil {
		panic(err)
	}

	var sessionRepo repositories.SessionRepository = implRepo.NewSqliteUserSessionRepository(db)
	var authRepo repositories.AuthRepository = implRepo.NewDtfAuthRepository(dtfService)
	var postRepo repositories.PostRepository = implRepo.NewDtfPostRepository(dtfService)

	var userManager managers.UserManager = implManagers.NewUserSessionManager(sessionRepo, authRepo)

	posts, err := FindRaffleNews(postRepo)
	if err != nil {
		log.Fatal(err)
	}
	for _, post := range posts {
		fmt.Println(post)
	}

	newUser, err := LoginUser(ctx, userManager, email, password)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("user is", newUser)
}

func FindRaffleNews(repo repositories.PostRepository) ([]models.Post, error) {
	uc := usecases.NewGetActiveRafflePostsUseCase(repo)
	posts, err := uc.Execute(time.Now().AddDate(0, 0, -1))
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func LoginUser(
	ctx context.Context,
	userManager managers.UserManager,
	email,
	password string,
) (models.UserSession, error) {
	user, err := userManager.EmailLogin(ctx, email, password)
	if err != nil {
		return models.UserSession{}, err
	}

	return user, nil
}
