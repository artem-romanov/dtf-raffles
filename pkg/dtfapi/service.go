// dtfapi provides methods to several DTF.RU apis,
// such as log in, refresh token, get post by id, search, etc
package dtfapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/guregu/null/v6"
	"resty.dev/v3"
)

const DtfApi string = "https://api.dtf.ru/"

type DtfService struct {
	client *resty.Client
}

func NewService(client *resty.Client) *DtfService {
	client.SetBaseURL(DtfApi)

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

func (c *DtfService) EmailLogin(ctx context.Context, email, password string) (Tokens, error) {
	var apiResult EmailLoginResponse
	var apiError DtfErrorV3

	resp, err := c.
		client.
		R().
		SetContext(ctx).
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
func (c *DtfService) RefreshToken(ctx context.Context, refreshToken string) (Tokens, error) {
	var apiResult RefreshTokenResponse
	var apiError DtfErrorV3

	resp, err := c.
		client.
		R().
		SetContext(ctx).
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

type SelfUserResponse struct {
	Message string `json:"message"`
	Result  struct {
		Id   int    `json:"id"`
		Url  string `json:"url"`
		Name string `json:"name"`
	}
}

// SelfUserInfo returns information about logged in user
func (c *DtfService) SelfUserInfo(ctx context.Context, accessToken string) (UserInfo, error) {
	var apiResponse SelfUserResponse
	var apiError DtfErrorV2

	req, err := c.withAuth(accessToken)
	if err != nil {
		return UserInfo{}, err
	}

	resp, err := req.
		SetContext(ctx).
		SetResult(&apiResponse).
		SetError(&apiError).
		Get("/v2.1/subsite/me")
	if err != nil {
		return UserInfo{}, err
	}

	if resp.IsError() {
		return UserInfo{}, err
	}

	return UserInfo{
		Id:   apiResponse.Result.Id,
		Url:  apiResponse.Result.Url,
		Name: apiResponse.Result.Name,
	}, nil
}

type PostBlock struct {
	Type   string          `json:"type"`
	Hidden bool            `json:"hidden"`
	Cover  bool            `json:"cover"`
	Data   json.RawMessage `json:"data"`
}

type PostResponse struct {
	Id       int         `json:"id"`
	Date     int         `json:"date"`
	Title    string      `json:"title"`
	Uri      string      `json:"url"`
	Blocks   []PostBlock `json:"blocks"`
	RepostId null.Int32  `json:"repostId"`
}

func (c *DtfService) GetPostById(
	ctx context.Context,
	id string,
) (BlogPost, error) {
	var apiError DtfErrorV2
	var apiResponse struct {
		Post PostResponse `json:"result"`
	}

	resp, err := c.client.
		R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"markdown": "false",
			"id":       id,
		}).
		SetResult(&apiResponse).
		SetError(&apiError).
		Get("/v2.10/content")
	if err != nil {
		return BlogPost{}, err
	} else if resp.IsError() {
		return BlogPost{}, apiError
	}

	blogPost, err := mapPostResponseToBlogPost(&apiResponse.Post)
	if err != nil {
		return BlogPost{}, err
	}

	return blogPost, nil
}

type SearchPostResponse struct {
	Result struct {
		Posts []struct {
			Data PostResponse `json:"data"`
		} `json:"items"`
	} `json:"result"`
}

func (c *DtfService) SearchNews(
	ctx context.Context,
	query string,
	dateFrom time.Time,
) ([]BlogPost, error) {
	var apiResponse SearchPostResponse
	var apiError DtfErrorV3

	resp, err := c.client.
		R().
		SetContext(ctx).
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
		blogPost, err := mapPostResponseToBlogPost(&apiPost.Data)
		if err != nil {
			// мы должны отправить хоть что-то любой ценой
			// logging only and continue
			slog.Error("blogpost unmarshal error", "err", err)
			continue
		}
		result = append(result, blogPost)
	}

	return result, nil
}

type ReactToPostRequest struct {
	Type int `json:"type"`
}

// Reacts to post with <Heart> reaction
func (c *DtfService) ReactToPost(
	ctx context.Context,
	accessToken string,
	postId int,
) error {
	var apiError DtfErrorV2

	req, err := c.withAuth(accessToken)
	if err != nil {
		return err
	}

	resp, err := req.
		SetContext(ctx).
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

func (c *DtfService) PostComment(ctx context.Context, accessToken string, postId int, text string) error {
	var apiError DtfErrorV2
	req, err := c.withAuth(accessToken)
	if err != nil {
		return err
	}

	resp, err := req.
		SetContext(ctx).
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

func mapPostResponseToBlogPost(response *PostResponse) (BlogPost, error) {
	var blocks []DataBlock

	for _, block := range response.Blocks {
		switch block.Type {
		case "text":
			var textBlock struct {
				Text string `json:"text"`
			}
			err := json.Unmarshal(block.Data, &textBlock)
			if err != nil {
				return BlogPost{}, err
			}
			blocks = append(blocks, DataText{
				HtmlText: textBlock.Text,
			})
		case "header":
			var headerBlock struct {
				Style string `json:"style"`
				Text  string `json:"text"`
			}
			err := json.Unmarshal(block.Data, &headerBlock)
			if err != nil {
				return BlogPost{}, err
			}
			blocks = append(blocks, DataHeader{
				Style: headerBlock.Style,
				Text:  headerBlock.Text,
			})
		case "list":
			var listBlock struct {
				Items []string `json:"items"`
				Type  string   `json:"type"`
			}
			err := json.Unmarshal(block.Data, &listBlock)
			if err != nil {
				return BlogPost{}, err
			}
			blocks = append(blocks, DataList{
				items: listBlock.Items,
			})
		}
		// TODO: Add other types when need comes
		// Also might be better idea to export this code into another function
	}

	return BlogPost{
		Id:        response.Id,
		Title:     response.Title,
		Uri:       response.Uri,
		Blocks:    blocks,
		RepliedTo: response.RepostId,
	}, nil
}
