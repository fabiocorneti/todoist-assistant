package todoist

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

const (
	maxRequests = 450
	interval    = 15 * time.Minute
)

// RateLimitedClient is a wrapper around http.Client that enforces rate limits.
type RateLimitedClient struct {
	client  *http.Client
	limiter *rate.Limiter
}

// NewRateLimitedClient creates a new RateLimitedClient.
func NewRateLimitedClient() *RateLimitedClient {
	rlc := &RateLimitedClient{
		client:  &http.Client{},
		limiter: rate.NewLimiter(rate.Every(interval/time.Duration(maxRequests)), maxRequests),
	}
	return rlc
}

// Do sends an HTTP request and returns an HTTP response, respecting the rate limit.
func (rlc *RateLimitedClient) Do(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	err := rlc.limiter.Wait(ctx)
	if err != nil {
		return nil, err
	}

	return rlc.client.Do(req)
}
