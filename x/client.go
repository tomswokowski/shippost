package x

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dghubble/oauth1"
	"github.com/tomswokowski/shippost/config"
)

const (
	// httpTimeout is the timeout for HTTP requests
	httpTimeout = 30 * time.Second
)

const (
	postsEndpoint  = "https://api.x.com/2/tweets"
	uploadEndpoint = "https://upload.twitter.com/1.1/media/upload.json"
	maxPostLength  = 280
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

// MediaResponse represents the API response for media upload
type MediaResponse struct {
	MediaID       int64  `json:"media_id"`
	MediaIDString string `json:"media_id_string"`
}

// APIError represents an error response from the X API
type APIError struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Type   string `json:"type"`
}

// PostOptions contains optional parameters for posting
type PostOptions struct {
	ReplyToID string   // ID of post to reply to (for threads)
	MediaIDs  []string // Media IDs to attach
}

// NewClient creates a new X API client with OAuth 1.0a authentication
func NewClient(cfg *config.Config) *Client {
	oauthConfig := oauth1.NewConfig(cfg.APIKey, cfg.APISecret)
	token := oauth1.NewToken(cfg.AccessToken, cfg.AccessSecret)
	httpClient := oauthConfig.Client(oauth1.NoContext, token)

	// Set timeout to prevent hanging on slow/unresponsive servers
	httpClient.Timeout = httpTimeout

	return &Client{
		httpClient: httpClient,
	}
}

// Post creates a new post on X
func (c *Client) Post(text string) (*PostResponse, error) {
	return c.PostWithOptions(text, nil)
}

// PostWithOptions creates a new post with additional options
func (c *Client) PostWithOptions(text string, opts *PostOptions) (*PostResponse, error) {
	// Validate post length using rune count for proper Unicode support
	runeCount := utf8.RuneCountInString(text)
	if runeCount == 0 {
		return nil, fmt.Errorf("post text cannot be empty")
	}
	if runeCount > maxPostLength {
		return nil, fmt.Errorf("post exceeds %d characters (%d)", maxPostLength, runeCount)
	}

	// Build request body
	body := map[string]interface{}{
		"text": text,
	}

	if opts != nil {
		if opts.ReplyToID != "" {
			body["reply"] = map[string]string{
				"in_reply_to_tweet_id": opts.ReplyToID,
			}
		}
		if len(opts.MediaIDs) > 0 {
			body["media"] = map[string]interface{}{
				"media_ids": opts.MediaIDs,
			}
		}
	}

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

// UploadMedia uploads an image or video and returns the media ID
func (c *Client) UploadMedia(filePath string) (*MediaResponse, error) {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Determine media type
	ext := strings.ToLower(filepath.Ext(filePath))
	var mediaType string
	switch ext {
	case ".jpg", ".jpeg":
		mediaType = "image/jpeg"
	case ".png":
		mediaType = "image/png"
	case ".gif":
		mediaType = "image/gif"
	case ".webp":
		mediaType = "image/webp"
	case ".mp4":
		mediaType = "video/mp4"
	default:
		return nil, fmt.Errorf("unsupported media type: %s", ext)
	}

	// For images, use simple upload
	if strings.HasPrefix(mediaType, "image/") {
		return c.uploadSimple(data, mediaType)
	}

	// For videos, would need chunked upload (not implemented yet)
	return nil, fmt.Errorf("video upload not yet supported")
}

// uploadSimple performs a simple media upload for images
func (c *Client) uploadSimple(data []byte, mediaType string) (*MediaResponse, error) {
	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add media_data field (base64 encoded)
	encoded := base64.StdEncoding.EncodeToString(data)
	if err := writer.WriteField("media_data", encoded); err != nil {
		return nil, fmt.Errorf("failed to write media data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", uploadEndpoint, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload media: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle errors
	if resp.StatusCode != http.StatusOK {
		return nil, parseAPIError(resp.StatusCode, respBody)
	}

	// Parse response
	var mediaResp MediaResponse
	if err := json.Unmarshal(respBody, &mediaResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &mediaResp, nil
}

// PostThread posts a series of connected posts as a thread
func (c *Client) PostThread(posts []ThreadPost) ([]*PostResponse, error) {
	if len(posts) == 0 {
		return nil, fmt.Errorf("thread cannot be empty")
	}

	var responses []*PostResponse
	var replyToID string

	for i, post := range posts {
		opts := &PostOptions{
			ReplyToID: replyToID,
			MediaIDs:  post.MediaIDs,
		}

		resp, err := c.PostWithOptions(post.Text, opts)
		if err != nil {
			return responses, fmt.Errorf("failed to post thread item %d: %w", i+1, err)
		}

		responses = append(responses, resp)
		replyToID = resp.Data.ID
	}

	return responses, nil
}

// ThreadPost represents a single post in a thread
type ThreadPost struct {
	Text     string
	MediaIDs []string
}

// parseAPIError extracts error details from API response
func parseAPIError(statusCode int, body []byte) error {
	var apiErr struct {
		Errors []APIError `json:"errors"`
		Title  string     `json:"title"`
		Detail string     `json:"detail"`
		Error  string     `json:"error"` // v1.1 API format
	}

	if err := json.Unmarshal(body, &apiErr); err != nil {
		return fmt.Errorf("API error (status %d)", statusCode)
	}

	// Check for v1.1 error format
	if apiErr.Error != "" {
		return fmt.Errorf("API error: %s", apiErr.Error)
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
