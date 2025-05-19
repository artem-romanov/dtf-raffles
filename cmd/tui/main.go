package main

import (
	"context"
	irepo "dtf/game_draw/internal/domain/repositories"
	"dtf/game_draw/internal/repositories"
	"dtf/game_draw/internal/storage/sqlite"
	"dtf/game_draw/internal/usecases"
	"dtf/game_draw/pkg/dtfapi"
	"errors"
	"fmt"
	"log"
	"time"
)

func main() {
	ctx := context.Background()
	postRepo, sessionRepo, err := initialize(ctx)
	if err != nil {
		log.Fatal(err)
	}

	user, err := sessionRepo.GetByEmail(ctx, "xenox2048@gmail.com")
	if errors.Is(err, repositories.ErrUserSessionNotFound) {
		log.Fatal("User not found")
	} else if err != nil {
		log.Fatal(err)
	}

	searchUC := usecases.NewGetActiveRafflePostsUseCase(postRepo)
	posts, err := searchUC.Execute(time.Now().AddDate(0, 0, -1))
	if err != nil {
		panic(err)
	}

	for _, post := range posts {
		fmt.Println(post)
	}
}

func initialize(ctx context.Context) (irepo.PostRepository, irepo.SessionRepository, error) {
	apiClient := dtfapi.NewClient(ctx)
	dtfService := dtfapi.NewService(apiClient.Client())

	db, err := sqlite.InitDB("./dtf_db.sqlite")
	if err != nil {
		return nil, nil, err
	}

	var sessionRepo irepo.SessionRepository = repositories.NewSqliteUserSessionRepository(db)
	var postRepo irepo.PostRepository = repositories.NewDtfPostRepository(dtfService)
	return postRepo, sessionRepo, nil
}

func authenticateUser(email string, password string) {

}
