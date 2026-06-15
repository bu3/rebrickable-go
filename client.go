package rebrickable

import (
	"fmt"
	nethttp "net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

const defaultBaseURL = "https://rebrickable.com/api/v3"

type Client struct {
	http      *resty.Client
	authToken string
}

func NewClient(apiKey string) *Client {
	return newClientWithBaseURL(apiKey, "", defaultBaseURL)
}

func NewAuthenticatedClient(apiKey, username, password string) (*Client, error) {
	c := newClientWithBaseURL(apiKey, "", defaultBaseURL)
	token, err := c.getUserToken(username, password)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}
	return newClientWithBaseURL(apiKey, token, defaultBaseURL), nil
}

func newClientWithBaseURL(apiKey, authToken, baseURL string) *Client {
	timeout := envDuration("REBRICKABLE_TIMEOUT", 30*time.Second)
	retryCount := envInt("REBRICKABLE_RETRY_COUNT", 3)
	retryWait := envDuration("REBRICKABLE_RETRY_WAIT", 10*time.Second)

	http := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("key %s", apiKey)).
		SetTimeout(timeout).
		SetRetryCount(retryCount).
		SetRetryWaitTime(retryWait).
		SetRetryMaxWaitTime(60 * time.Second).
		AddRetryCondition(func(r *resty.Response, _ error) bool {
			return r != nil && r.StatusCode() == nethttp.StatusTooManyRequests
		})

	return &Client{http: http, authToken: authToken}
}

func envDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}

func envInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}

func (c *Client) getUserToken(username, password string) (string, error) {
	type tokenResponse struct {
		UserToken string `json:"user_token"`
	}
	result := &tokenResponse{}
	resp, err := c.http.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{"username": username, "password": password}).
		SetResult(result).
		Post("/users/_token/")
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("token request failed with status %d", resp.StatusCode())
	}
	return result.UserToken, nil
}

func (c *Client) userPath(path string) string {
	return fmt.Sprintf("/users/%s%s", c.authToken, path)
}

func fetchAllPages[T any](httpClient *resty.Client, firstURL string) (int, []T, error) {
	type page struct {
		Count   int    `json:"count"`
		Next    string `json:"next"`
		Results []T    `json:"results"`
	}

	var all []T
	totalCount := 0
	isFirst := true
	url := firstURL

	for {
		p := &page{}
		resp, err := httpClient.R().SetResult(p).Get(url)
		if err != nil {
			return 0, nil, fmt.Errorf("request failed: %w", err)
		}
		if resp.StatusCode() != 200 {
			return 0, nil, fmt.Errorf("unexpected status %d", resp.StatusCode())
		}
		if isFirst {
			totalCount = p.Count
			isFirst = false
		}
		all = append(all, p.Results...)
		if p.Next == "" {
			break
		}
		url = p.Next
	}
	return totalCount, all, nil
}
