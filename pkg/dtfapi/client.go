package dtfapi

import (
	"context"
	"errors"
	"math/rand/v2"

	"golang.org/x/time/rate"
	"resty.dev/v3"
)

type apiClient struct {
	client  *resty.Client
	limiter *rate.Limiter
}

var userAgents = []string{
	"Mozilla/5.0 (iPhone; CPU iPhone OS 12_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/86.0.4240.93 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Linux; Android 9; Symphony G10) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.92 Mobile Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.183 Safari/537.36 OPR/72.0.3815.320",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/81.0.4041.0 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Linux; Android 10; SM-A217F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.86 Mobile Safari/537.36",
	"Mozilla/5.0 (Linux; Android 10; EML-L29) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.85 Mobile Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4185.0 Safari/537.36",
	"Mozilla/5.0 (Linux; Android 10; SAMSUNG SM-A207M) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/11.0 Chrome/75.0.3770.143 Mobile Safari/537.36",
	"Mozilla/5.0 (Linux; Android 6.0.1; Le X527 Build/IMXOSOP5801910311S) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/49.0.2623.91 Mobile Safari/537.36 YaApp_Android/10.61 YaSearchBrowser/10.61",
	"Mozilla/5.0 (Linux; Android 10; Infinix X683 Build/QP1A.190711.020; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/89.0.4389.86 Mobile Safari/537.36",
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
	c.client.AddRequestMiddleware(func(c *resty.Client, r *resty.Request) error {
		userAgent := getRandomUserAgent()
		r.SetHeader("User-agent", userAgent)
		return nil
	})
	c.client.AddRequestMiddleware(func(_ *resty.Client, r *resty.Request) error {
		err := c.limiter.Wait(r.Context())
		if err != nil {
			return err
		}

		// uncomment if doubt about rate limiting
		// TODO: remove in the nearest future
		// slog.Info("request allowed at", "time", time.Now())

		return nil
	})
}

func getRandomUserAgent() string {
	randomIndex := rand.IntN(len(userAgents))
	return userAgents[randomIndex]
}
