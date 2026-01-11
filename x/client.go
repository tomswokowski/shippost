package x

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dghubble/oauth1"
	"github.com/tom/shippost/config"
)

const (
	postsEndpoint = "https://api.x.com/2/tweets"
	maxPostLength = 280
)

// Client handles X API interactions
type Client struct {
	httpClient *http.Client
}

// PostResponse represents the API response for creating a post
type PostResponse struct {
	Data struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	} `json:"data"`
}

// APIError represents an error response from the X API
type APIError struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Type   string `json:"type"`
}

// NewClient creates a new X API client with OAuth 1.0a authentication
func NewClient(cfg *config.Config) *Client {
	oauthConfig := oauth1.NewConfig(cfg.APIKey, cfg.APISecret)
	token := oauth1.NewToken(cfg.AccessToken, cfg.AccessSecret)
	httpClient := oauthConfig.Client(oauth1.NoContext, token)

	return &Client{
		httpClient: httpClient,
	}
}

// Post creates a new post on X
func (c *Client) Post(text string) (*PostResponse, error) {
	// Validate post length
	if len(text) == 0 {
		return nil, fmt.Errorf("post text cannot be empty")
	}
	if len(text) > maxPostLength {
		return nil, fmt.Errorf("post exceeds %d characters (%d)", maxPostLength, len(text))
	}

	// Prepare request body
	body := map[string]string{"text": text}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", postsEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle errors
	if resp.StatusCode != http.StatusCreated {
		return nil, parseAPIError(resp.StatusCode, respBody)
	}

	// Parse success response
	var postResp PostResponse
	if err := json.Unmarshal(respBody, &postResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &postResp, nil
}

// parseAPIError extracts error details from API response
func parseAPIError(statusCode int, body []byte) error {
	var apiErr struct {
		Errors []APIError `json:"errors"`
		Title  string     `json:"title"`
		Detail string     `json:"detail"`
	}

	if err := json.Unmarshal(body, &apiErr); err != nil {
		return fmt.Errorf("API error (status %d)", statusCode)
	}

	// Check for errors array
	if len(apiErr.Errors) > 0 {
		return fmt.Errorf("API error: %s", apiErr.Errors[0].Detail)
	}

	// Check for direct error fields
	if apiErr.Detail != "" {
		return fmt.Errorf("API error: %s", apiErr.Detail)
	}

	if apiErr.Title != "" {
		return fmt.Errorf("API error: %s", apiErr.Title)
	}

	return fmt.Errorf("API error (status %d)", statusCode)
}
