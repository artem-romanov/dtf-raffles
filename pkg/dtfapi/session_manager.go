package dtfapi

import (
	"errors"
	"fmt"

	"resty.dev/v3"
)

type SessionManager struct {
}

func NewSessionManager() *SessionManager {
	return &SessionManager{}
}

func (sm *SessionManager) Authenticate(email string, tokens Tokens) (UserSession, error) {
	if !tokens.IsAccessValid() {
		return UserSession{
			Email:      email,
			UserTokens: tokens,
		}, nil
	}

	// trying to refresh
	sm.RefreshTokens()
	return UserSession{}, errors.New("NOT IMPLEMENTED")
}

func (sm *SessionManager) RefreshTokens() (*resty.Request, error) {
	authenticated := c.userSession.Authenticated
	tokens := c.userSession.UserTokens

	if !authenticated {
		return nil, errors.New("User is not authenticated")
	}

	if tokens.AccessToken == "" {
		return nil, ErrMissingToken
	}

	if !tokens.IsAccessValid() {
		// fuck, looks like we need to update tokens
		// singleflight helps for concurrency
		// 	and avoids race conditions (N requests for refresh = only 1 correct refresh for all)
		tokensAny, err, _ := c.refreshTokenGroup.Do("refresh", func() (interface{}, error) {
			return c.RefreshToken(tokens.RefreshToken)
		})
		if err != nil {
			return nil, fmt.Errorf("Couldn't refresh tokens: %s", err)
		}
		if newTokens, ok := tokensAny.(Tokens); ok {
			c.setTokens(newTokens)
		} else {
			return nil, fmt.Errorf("Can't cast Tokens from refresh token [withAuth]")
		}
	}

	// finally, taking all tokens again
	// in case if they were updated
	c.userSessionMu.RLock()
	tokens = c.userSession.UserTokens
	c.userSessionMu.RUnlock()

	headerValue := fmt.Sprintf("Bearer %s", tokens.AccessToken)
	return c.client.R().SetHeader("Jwtauthorization", headerValue), nil
}
