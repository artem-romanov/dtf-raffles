package managers

import (
	"context"
	"dtf/game_draw/internal/domain"
	"dtf/game_draw/internal/domain/models"
	"dtf/game_draw/internal/domain/repositories"
	"errors"
	"log/slog"
	"time"
)

type userSessionManager struct {
	sessionRepo repositories.SessionRepository
	authRepo    repositories.AuthRepository
}

func NewUserSessionManager(
	sessionRepo repositories.SessionRepository,
	authRepo repositories.AuthRepository,
) *userSessionManager {
	return &userSessionManager{
		sessionRepo: sessionRepo,
		authRepo:    authRepo,
	}
}

func (usm *userSessionManager) BuildSession(ctx context.Context, email string) (models.UserSession, error) {
	user, err := usm.sessionRepo.GetByEmail(ctx, email)
	if err != nil {
		return models.UserSession{}, err
	}

	if !UserExpired(user) {
		return user, nil
	}

	newUser, err := usm.authRepo.RefreshToken(user)
	if err != nil {
		return models.UserSession{}, err
	}

	// Right now it's ok to skip if error
	// TODO: test and think about it later
	usm.persistUser(ctx, newUser)
	return newUser, nil
}

func (usm *userSessionManager) EmailLogin(ctx context.Context, email, password string) (models.UserSession, error) {
	user, err := usm.authRepo.Login(email, password)
	if err == nil {
		// Right now it's ok to skip if error
		// TODO: test and think about it later
		usm.persistUser(ctx, user)
	}

	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			deleteErr := usm.sessionRepo.DeleteByEmail(ctx, email)
			if deleteErr != nil {
				return models.UserSession{}, deleteErr
			}
		}
		return models.UserSession{}, err
	}

	return user, nil
}

func (usm *userSessionManager) persistUser(ctx context.Context, user models.UserSession) error {
	err := usm.sessionRepo.Save(ctx, user)
	if err != nil {
		slog.Error("Couldnt save updated user", "email", user.Email)
		return err
	}
	return nil
}

// TODO: Move to utils or domain model support functions
func UserExpired(session models.UserSession) bool {
	if session.AccessToken == "" {
		return true
	}

	diff := time.Until(session.AccessExpiration)
	if diff.Microseconds() <= 0 {
		return true
	}

	return false
}
