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

func (r *dtfAuthRepository) Login(email, password string) (models.DtfUserSession, error) {
	tokens, err := r.dtfService.EmailLogin(email, password)

	if err != nil {
		if errors.Is(err, dtfapi.ErrInvalidCredentials) {
			return models.DtfUserSession{}, domain.ErrInvalidCredentials
		}
		return models.DtfUserSession{}, err
	}

	return models.DtfUserSession{
		Email:            email,
		AccessToken:      tokens.AccessToken,
		RefreshToken:     tokens.RefreshToken,
		AccessExpiration: tokens.AccessExpiration,
	}, nil

}

func (r *dtfAuthRepository) RefreshToken(user models.DtfUserSession) (models.DtfUserSession, error) {
	tokens, err := r.dtfService.RefreshToken(user.RefreshToken)
	if err != nil {
		return models.DtfUserSession{}, err
	}

	return models.DtfUserSession{
		Email:            user.Email,
		AccessToken:      tokens.AccessToken,
		RefreshToken:     tokens.RefreshToken,
		AccessExpiration: tokens.AccessExpiration,
	}, nil
}

func (r *dtfAuthRepository) SelfInfo(user models.DtfUserSession) (models.DtfUserInfo, error) {
	response, err := r.dtfService.SelfUserInfo(user.AccessToken)
	if err != nil {
		return models.DtfUserInfo{}, err
	}

	return models.DtfUserInfo{
		Id:   response.Id,
		Name: response.Name,
		Url:  response.Url,
	}, nil
}
