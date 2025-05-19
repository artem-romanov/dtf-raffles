package main

import (
	"context"
	"dtf/game_draw/internal/domain/models"
	repo "dtf/game_draw/internal/domain/repositories"
	"dtf/game_draw/internal/repositories"
	"dtf/game_draw/internal/storage/sqlite"
	"dtf/game_draw/pkg/dtfapi"
	"errors"
	"fmt"
)

const (
	email      string = "xenox2048@gmail.com"
	pass       string = "Azxsdcvf12"
	sqlitePath string = "./dtf_db.sqlite"
)

func main() {
	ctx := context.Background()

	db, err := sqlite.InitDB(sqlitePath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	client := dtfapi.NewClient(context.TODO())
	defer client.Close()

	dtfService := dtfapi.NewService(client.Client())

	var userSessionRepo repo.SessionRepository = repositories.NewSqliteUserSessionRepository(db)
	var postRepository repo.PostRepository = repositories.NewDtfPostRepository(dtfService)

	dtfService.SetOnRefreshCallback(func(email string, tokens dtfapi.Tokens) {
		// TODO: check sqlite and think about mutex
		session := models.UserSession{
			Email:            email,
			AccessToken:      tokens.AccessToken,
			RefreshToken:     tokens.RefreshToken,
			AccessExpiration: tokens.AccessExpiration,
		}
		err := userSessionRepo.Save(ctx, session)

		if err != nil {
			fmt.Println("COULDNT SAVE SESSION: ", session.Email)
		}
	})
	session, err := findExistingSession(ctx, userSessionRepo, email)
	if err == nil {
		fmt.Println("USER HAS BEEN FOUND!")
	}

	if err != nil && !errors.Is(err, repositories.ErrUserSessionNotFound) {
		panic(err)
	}

	if err != nil && errors.Is(err, repositories.ErrUserSessionNotFound) {
		session, err = authenticateUser(dtfService, email, pass)
		if err != nil {
			panic(err)
		}
		err = userSessionRepo.Save(ctx, session)
		if err != nil {
			panic(err)
		}
	}

	dtfService.SetSession(email, dtfapi.Tokens{
		AccessToken:      session.AccessToken,
		RefreshToken:     session.RefreshToken,
		AccessExpiration: session.AccessExpiration,
	})

	err = postRepository.ReactToPost(models.Post{
		Id: 3772402,
	})
	if err != nil {
		panic(err)
	}
}

func findExistingSession(
	ctx context.Context,
	sessionRepo repo.SessionRepository,
	email string,
) (models.UserSession, error) {
	// check if session exists in DB
	session, err := sessionRepo.GetByEmail(ctx, email)
	if err != nil {
		return models.UserSession{}, err
	}

	return session, nil
}

func authenticateUser(
	dtfService *dtfapi.DtfService,
	email string,
	password string,
) (models.UserSession, error) {
	fmt.Println("TRYING TO LOGIN USER")
	tokens, err := dtfService.EmailLogin(email, password)
	if err != nil {
		return models.UserSession{}, err
	}

	return models.UserSession{
		Email:            email,
		AccessToken:      tokens.AccessToken,
		RefreshToken:     tokens.RefreshToken,
		AccessExpiration: tokens.AccessExpiration,
	}, nil
}
