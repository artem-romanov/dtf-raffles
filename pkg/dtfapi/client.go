package dtfapi

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"golang.org/x/time/rate"
	"resty.dev/v3"
)

type apiClient struct {
	client  *resty.Client
	limiter *rate.Limiter
}

func NewClient(ctx context.Context) *apiClient {
	limiter := rate.NewLimiter(rate.Limit(1/3.0), 3) // we allow only 3 request per second
	restClient := resty.New()

	result := &apiClient{
		client:  restClient,
		limiter: limiter,
	}
	result.initClientMiddlewares(ctx)
	return result
}

func (c apiClient) Client() *resty.Client {
	return c.client
}

func (c *apiClient) Close() error {
	if c.client == nil {
		return errors.New("Already closed")
	}

	return c.client.Close()
}

func (c *apiClient) initClientMiddlewares(ctx context.Context) {
	c.client.AddRequestMiddleware(func(_ *resty.Client, _ *resty.Request) error {
		err := c.limiter.Wait(ctx)
		if err != nil {
			return err
		}
		slog.Info("request allowed at", "time", time.Now())
		return nil
	})
}
