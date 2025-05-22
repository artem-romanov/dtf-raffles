package main

import (
	"context"
	"dtf/game_draw/internal/domain/managers"
	"dtf/game_draw/internal/domain/models"
	"dtf/game_draw/internal/domain/repositories"
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
	defer db.Close()

	// var sessionRepo repositories.DtfSessionRepository = implRepo.NewSqliteUserSessionRepository(db)
	// var authRepo repositories.AuthRepository = implRepo.NewDtfAuthRepository(dtfService)
	var postRepo repositories.PostRepository = implRepo.NewDtfPostRepository(dtfService)

	// var userManager managers.UserManager = implManagers.NewUserSessionManager(sessionRepo, authRepo)

	posts, err := FindRaffleNews(ctx, postRepo)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Список Розыгрышей на сегодня:")
	for _, post := range posts {
		fmt.Printf("<%d>: %s\n (%s)\n[Description]: %s", post.Id, post.Title, post.Uri, post.Text)
	}
	fmt.Println("\n________")

	// slog.Info("Starting concurrent requests...")
	// var wg sync.WaitGroup
	// for i := range 10 {
	// 	wg.Add(1)
	// 	go func(number int) {
	// 		defer wg.Done()

	// 		session, err := userManager.BuildSession(ctx, email)
	// 		if err != nil {
	// 			fmt.Println("Error is: ", err)
	// 			return
	// 		}

	// 		userInfo, err := authRepo.SelfInfo(ctx, session)
	// 		if err != nil {
	// 			fmt.Println("Error userInfo: ", err)
	// 			return
	// 		}

	// 		fmt.Println(userInfo)
	// 	}(i + 1)
	// }

	// wg.Wait()
	// slog.Info("Exiting programm...")
}

func FindRaffleNews(ctx context.Context, repo repositories.PostRepository) ([]models.Post, error) {
	uc := usecases.NewGetActiveRafflePostsUseCase(repo)
	posts, err := uc.Execute(ctx, time.Now().AddDate(0, 0, -1))
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
) (models.DtfUserSession, error) {
	user, err := userManager.EmailLogin(ctx, email, password)
	if err != nil {
		return models.DtfUserSession{}, err
	}

	return user, nil
}
