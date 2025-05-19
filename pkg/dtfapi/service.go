package dtfapi

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
	"resty.dev/v3"
)

const DtfApi string = "https://api.dtf.ru/"

type DtfService struct {
	// private fields
	client        *resty.Client
	userSession   UserSession
	userSessionMu sync.RWMutex

	// utilities
	refreshTokenGroup singleflight.Group

	// callbacks
	onTokenRefresh func(email string, tokens Tokens)
}

func NewService(client *resty.Client) *DtfService {
	client.SetBaseURL(DtfApi)

	chromeUserAgent := "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"
	client.SetHeader("User-agent", chromeUserAgent)

	return &DtfService{
		client: client,
		userSession: UserSession{
			Email:      "",
			UserTokens: Tokens{},
		},
	}
}

func (c *DtfService) SetSession(email string, tokens Tokens) {
	c.userSessionMu.Lock()
	defer c.userSessionMu.Unlock()

	c.userSession = UserSession{
		Email:         email,
		UserTokens:    tokens,
		Authenticated: true,
	}
}

func (c *DtfService) DisposeSession() {
	c.userSessionMu.Lock()
	defer c.userSessionMu.Unlock()

	c.userSession = UserSession{
		Authenticated: false,
	}
}

type EmailLoginResponse struct {
	Data struct {
		AccessToken         string `json:"accessToken"`
		AccessExpTimestamp  int64  `json:"accessExpTimestamp"`
		RefreshToken        string `json:"refreshToken"`
		RefreshExpTimestamp int64  `json:"refreshExpTimestamp"`
	} `json:"data"`
}

func (c *DtfService) EmailLogin(email, password string) (Tokens, error) {
	var apiResult EmailLoginResponse
	var apiError DtfErrorV3

	resp, err := c.
		client.
		R().
		SetMultipartFormData(map[string]string{
			"email":    email,
			"password": password,
		}).
		SetResult(&apiResult).
		SetError(&apiError).
		Post("/v3.0/auth/email/login")

	if err != nil {
		return Tokens{}, err
	}

	if resp.IsError() {
		return Tokens{}, apiError
	}

	return Tokens{
		AccessToken:      apiResult.Data.AccessToken,
		RefreshToken:     apiResult.Data.RefreshToken,
		AccessExpiration: time.Unix(apiResult.Data.AccessExpTimestamp, 0),
	}, nil
}

type RefreshTokenRequest struct {
	Token string `json:"token"`
}

type RefreshTokenResponse struct {
	Message string `json:"message"`
	Data    struct {
		AccessToken         string `json:"accessToken"`
		AccessExpTimestamp  int64  `json:"accessExpTimestamp"`
		RefreshToken        string `json:"refreshToken"`
		RefreshExpTimestamp int64  `json:"refreshExpTimestamp"`
	} `json:"data"`
}

// NOTE: because it refresh token operation, it will dispose previous pair
// This method doesn't update user session
func (c *DtfService) RefreshToken(refreshToken string) (Tokens, error) {
	var apiResult RefreshTokenResponse
	var apiError DtfErrorV3

	resp, err := c.
		client.
		R().
		SetResult(&apiResult).
		SetError(&apiError).
		SetMultipartFormData(map[string]string{
			"token": refreshToken,
		}).
		Post("/v3.0/auth/refresh")

	if err != nil {
		return Tokens{}, err
	}

	if resp.IsError() {
		return Tokens{}, apiError
	}

	if apiResult.Message == "Refresh token is missing" {
		return Tokens{}, DtfErrorV3{
			Message: "Refresh token is missing",
			Code:    400, // this is a lie, fucking api sends 200. but i dont care.
		}
	}

	tokens := Tokens{
		AccessToken:      apiResult.Data.AccessToken,
		RefreshToken:     apiResult.Data.RefreshToken,
		AccessExpiration: time.Unix(apiResult.Data.AccessExpTimestamp, 0),
	}

	return tokens, nil
}

type PostResponse struct {
	Data struct {
		Id    int    `json:"id"`
		Date  int    `json:"date"`
		Title string `json:"title"`
		Uri   string `json:"url"`
	} `json:"data"`
}

type SearchPostResponse struct {
	Result struct {
		Posts []PostResponse `json:"items"`
	} `json:"result"`
}

func (c *DtfService) SearchNews(query string, dateFrom time.Time) ([]BlogPost, error) {
	var apiResponse SearchPostResponse
	var apiError DtfErrorV3

	resp, err := c.client.
		R().
		SetQueryParams(map[string]string{
			"markdown":  "false",
			"sorting":   "date",
			"q":         query,
			"title":     "true",
			"editorial": "false",
			"strict":    "false",
			"dateFrom":  fmt.Sprintf("%d", dateFrom.Unix()),
		}).
		SetResult(&apiResponse).
		SetError(&apiError).
		Get("/v2.8/search/posts")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, apiError
	}

	result := make([]BlogPost, 0, len(apiResponse.Result.Posts))

	for _, apiPost := range apiResponse.Result.Posts {
		result = append(result, BlogPost{
			Id:    apiPost.Data.Id,
			Title: apiPost.Data.Title,
			Uri:   apiPost.Data.Uri,
		})
	}

	return result, nil
}

type ReactToPostRequest struct {
	Type int `json:"type"`
}

// Reacts to post with <Heart> reaction
func (c *DtfService) ReactToPost(postId int) error {
	var apiError DtfErrorV2

	req, err := c.withAuth()

	if err != nil {
		return err
	}

	resp, err := req.
		SetError(&apiError).
		SetMultipartFormData(map[string]string{
			"type": strconv.Itoa(1), // This is the HEART reaction Id (id == 1)
		}).
		SetPathParam("post_id", strconv.Itoa(postId)).
		Post("/v2.5/content/{post_id}/react")

	if err != nil {
		return err
	}

	if resp.IsError() {
		return apiError
	}

	return nil
}

type PostCommentRequest struct {
	Id   int    `json:"id"`
	Text string `json:"text"`

	// User's id.
	// This can be used if we need to comment someone.
	ReplyTo int `json:"reply_to"`
	// we won't use it in real world,
	// but it needs to be described for proper API call.
	// IINM this is a list of url links or ids
	Attachments []string `json:"attachments"`
}

func (c *DtfService) PostComment(postId int, text string) error {
	var apiError DtfErrorV2
	req, err := c.withAuth()
	if err != nil {
		return err
	}

	resp, err := req.
		SetMultipartFormData(map[string]string{
			"id":          strconv.Itoa(postId),
			"text":        text,
			"reply_to":    "0",  // we don't need to reply anyone, just post under the blogpost
			"attachments": "[]", // providing empty attachment list, no images
		}).
		SetError(&apiError).
		Post("/v2.4/comment/add")

	if err != nil {
		return err
	}

	if resp.IsError() {
		return err
	}

	return nil
}

func (c *DtfService) GetTokens() Tokens {
	return c.userSession.UserTokens
}

func (c *DtfService) SetOnRefreshCallback(cb func(email string, tokens Tokens)) {
	c.onTokenRefresh = cb
}

func (c *DtfService) withAuth() (*resty.Request, error) {
	c.userSessionMu.RLock()
	authenticated := c.userSession.Authenticated
	tokens := c.userSession.UserTokens
	c.userSessionMu.RUnlock()

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
			fmt.Println(err)
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

func (c *DtfService) setTokens(tokens Tokens) {
	c.userSessionMu.Lock()
	c.userSession.UserTokens = tokens
	c.userSessionMu.Unlock()

	if c.onTokenRefresh != nil {
		c.onTokenRefresh(c.userSession.Email, tokens)
	}
}
