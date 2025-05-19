package dtfapi

import (
	"fmt"
	"strconv"
	"time"

	"resty.dev/v3"
)

const DtfApi string = "https://api.dtf.ru/"

type DtfService struct {
	client *resty.Client
}

func NewService(client *resty.Client) *DtfService {
	client.SetBaseURL(DtfApi)

	// TODO: Introduce 10 different user agents and pick randomly
	chromeUserAgent := "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"
	client.SetHeader("User-agent", chromeUserAgent)

	return &DtfService{
		client: client,
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
		if apiError.Code == 104 && apiError.Message == "Invalid login or password" {
			return Tokens{}, ErrInvalidCredentials
		}

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
func (c *DtfService) ReactToPost(accessToken string, postId int) error {
	var apiError DtfErrorV2

	req, err := c.withAuth(accessToken)

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

func (c *DtfService) PostComment(accessToken string, postId int, text string) error {
	var apiError DtfErrorV2
	req, err := c.withAuth(accessToken)
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

func (c *DtfService) withAuth(accessToken string) (*resty.Request, error) {
	headerValue := fmt.Sprintf("Bearer %s", accessToken)
	return c.client.R().SetHeader("Jwtauthorization", headerValue), nil
}
