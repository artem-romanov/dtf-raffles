package managers

import (
	"context"
	"dtf/game_draw/internal/domain"
	"dtf/game_draw/internal/domain/models"
	"dtf/game_draw/internal/domain/repositories"
	"dtf/game_draw/internal/utils"
	"errors"
	"fmt"
	"log/slog"

	"golang.org/x/sync/singleflight"
)

type userSessionManager struct {
	sessionRepo  repositories.DtfSessionRepository
	authRepo     repositories.AuthRepository
	refreshGroup singleflight.Group
}

func NewUserSessionManager(
	sessionRepo repositories.DtfSessionRepository,
	authRepo repositories.AuthRepository,
) *userSessionManager {
	return &userSessionManager{
		sessionRepo: sessionRepo,
		authRepo:    authRepo,
	}
}

func (usm *userSessionManager) BuildSession(ctx context.Context, email string) (models.DtfUserSession, error) {
	user, err := usm.sessionRepo.GetByEmail(ctx, email)
	if err != nil {
		return models.DtfUserSession{}, err
	}

	if !utils.UserExpired(user) {
		return user, nil
	}

	// Avoid refresh token race conditions
	// WARNING side effect: 1st executor will also place data in session repo
	// other gorutines will only receive results
	newUserAny, err, _ := usm.refreshGroup.Do(email, func() (interface{}, error) {
		slog.Info("Running singleflight", "user", user.Email)
		newUser, err := usm.authRepo.RefreshToken(ctx, user)
		if err != nil {
			return models.DtfUserSession{}, err
		}

		if err := usm.persistUser(ctx, user); err != nil {
			slog.Error("Failed to persist user", "email", user.Email, "error", err)
			return models.DtfUserSession{}, err
		}
		return newUser, nil
	})
	if err != nil {
		return models.DtfUserSession{}, err
	}

	newUser, ok := newUserAny.(models.DtfUserSession)
	if !ok {
		return models.DtfUserSession{}, fmt.Errorf("invalid type assertion for user session")
	}
	return newUser, nil
}

func (usm *userSessionManager) EmailLogin(ctx context.Context, email, password string) (models.DtfUserSession, error) {
	user, err := usm.authRepo.Login(ctx, email, password)
	if err == nil {
		// Right now it's ok to skip if error
		// TODO: test and think about it later
		usm.persistUser(ctx, user)
	}

	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			deleteErr := usm.sessionRepo.DeleteByEmail(ctx, nil, email)
			if deleteErr != nil {
				return models.DtfUserSession{}, deleteErr
			}
		}
		return models.DtfUserSession{}, err
	}

	return user, nil
}

func (usm *userSessionManager) persistUser(ctx context.Context, user models.DtfUserSession) error {
	slog.Info("Running persist", "user", user.Email)
	err := usm.sessionRepo.Save(ctx, nil, user)
	if err != nil {
		slog.Error("Couldnt save updated user", "email", user.Email)
		return err
	}
	return nil
}
