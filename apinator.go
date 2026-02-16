package apinator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is an Apinator API client for server-side use.
type Client struct {
	appID  string
	key    string
	secret string
	host   string
	http   *http.Client
}

// Option is a functional option for configuring the Client.
type Option func(*Client)

// withBaseURL overrides the host URL (for testing only).
func withBaseURL(url string) Option {
	return func(c *Client) {
		c.host = url
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.http = httpClient
	}
}

// NewClient creates a new Apinator API client.
// The cluster parameter identifies the data plane region (e.g. "eu", "us").
// The HTTP base URL is derived as https://ws-{cluster}.apinator.io.
func NewClient(appID, key, secret, cluster string, opts ...Option) *Client {
	c := &Client{
		appID:  appID,
		key:    key,
		secret: secret,
		host:   fmt.Sprintf("https://ws-%s.apinator.io", cluster),
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Trigger sends an event to one or more channels.
// Either Channel or Channels must be provided (not both).
func (c *Client) Trigger(ctx context.Context, params TriggerParams) error {
	path := fmt.Sprintf("/apps/%s/events", c.appID)

	body, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}

	var result map[string]interface{}
	if err := c.request(ctx, "POST", path, body, &result); err != nil {
		return err
	}

	return nil
}

// AuthenticateChannel generates an authentication response for a channel subscription.
// This is a convenience method that delegates to the AuthenticateChannel function.
func (c *Client) AuthenticateChannel(socketID, channelName string, channelData *string) ChannelAuthResponse {
	return AuthenticateChannel(c.secret, c.key, socketID, channelName, channelData)
}

// GetChannels retrieves a list of channels, optionally filtered by prefix.
func (c *Client) GetChannels(ctx context.Context, prefix string) ([]ChannelInfo, error) {
	path := fmt.Sprintf("/apps/%s/channels", c.appID)
	if prefix != "" {
		path += "?filter_by_prefix=" + url.QueryEscape(prefix)
	}

	var result struct {
		Channels []ChannelInfo `json:"channels"`
	}

	if err := c.request(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}

	return result.Channels, nil
}

// GetChannel retrieves information about a specific channel.
// Returns nil if the channel does not exist.
func (c *Client) GetChannel(ctx context.Context, channelName string) (*ChannelInfo, error) {
	path := fmt.Sprintf("/apps/%s/channels/%s", c.appID, url.PathEscape(channelName))

	var result ChannelInfo
	if err := c.request(ctx, "GET", path, nil, &result); err != nil {
		// Return nil if channel not found (404)
		if apiErr, ok := err.(*ApiError); ok && apiErr.Status == 404 {
			return nil, nil
		}
		return nil, err
	}

	return &result, nil
}

// VerifyWebhook verifies a webhook signature.
// This is a convenience method that delegates to the VerifyWebhook function.
func (c *Client) VerifyWebhook(headers http.Header, body []byte, maxAge time.Duration) bool {
	return VerifyWebhook(c.secret, headers, body, maxAge)
}

// request performs an authenticated HTTP request to the Apinator API.
func (c *Client) request(ctx context.Context, method, path string, body []byte, result interface{}) error {
	// Construct full URL
	fullURL := c.host + path

	// Create request
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Compute HMAC signature
	timestamp := time.Now().Unix()
	signPath := path
	if idx := strings.Index(path, "?"); idx >= 0 {
		signPath = path[:idx]
	}
	signature := SignRequest(c.secret, method, signPath, body, timestamp)

	// Set auth headers
	req.Header.Set("X-Realtime-Key", c.key)
	req.Header.Set("X-Realtime-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("X-Realtime-Signature", signature)

	// Perform request
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode >= 400 {
		return c.handleError(resp.StatusCode, respBody)
	}

	// Parse response
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// handleError creates an appropriate error type based on status code.
func (c *Client) handleError(status int, body []byte) error {
	var problem struct {
		Type   string `json:"type"`
		Title  string `json:"title"`
		Status int    `json:"status"`
		Detail string `json:"detail"`
		Code   string `json:"code"`
	}

	message := "API request failed"
	if err := json.Unmarshal(body, &problem); err == nil {
		if problem.Detail != "" {
			message = problem.Detail
		} else if problem.Title != "" {
			message = problem.Title
		}
	}

	bodyStr := string(body)

	switch status {
	case 401, 403:
		return &AuthenticationError{
			Message: message,
			Status:  status,
			Body:    bodyStr,
		}
	case 400, 422:
		return &ValidationError{
			Message: message,
			Status:  status,
			Body:    bodyStr,
		}
	default:
		return &ApiError{
			Message: message,
			Status:  status,
			Body:    bodyStr,
		}
	}
}
