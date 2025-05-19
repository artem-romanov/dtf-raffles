package repositories

import (
	"dtf/game_draw/internal/domain"
	"dtf/game_draw/internal/domain/models"
	"dtf/game_draw/pkg/dtfapi"
	"errors"
)

type dtfAuthRepository struct {
	dtfService *dtfapi.DtfService
}

func NewDtfAuthRepository(dtfService *dtfapi.DtfService) *dtfAuthRepository {
	return &dtfAuthRepository{
		dtfService: dtfService,
	}
}

func (r *dtfAuthRepository) Login(email, password string) (models.UserSession, error) {
	tokens, err := r.dtfService.EmailLogin(email, password)

	if err != nil {
		if errors.Is(err, dtfapi.ErrInvalidCredentials) {
			return models.UserSession{}, domain.ErrInvalidCredentials
		}
		return models.UserSession{}, err
	}

	return models.UserSession{
		Email:            email,
		AccessToken:      tokens.AccessToken,
		RefreshToken:     tokens.RefreshToken,
		AccessExpiration: tokens.AccessExpiration,
	}, nil

}

func (r *dtfAuthRepository) RefreshToken(user models.UserSession) (models.UserSession, error) {
	tokens, err := r.dtfService.RefreshToken(user.RefreshToken)
	if err != nil {
		return models.UserSession{}, err
	}

	return models.UserSession{
		Email:            user.Email,
		AccessToken:      tokens.AccessToken,
		RefreshToken:     tokens.RefreshToken,
		AccessExpiration: tokens.AccessExpiration,
	}, nil
}
