// Package proxy provides HTTP client for upstream API calls.
package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// Client is an HTTP client for making upstream API calls.
type Client struct {
	baseURL    string
	httpClient *http.Client
	headers    map[string]string
}

// NewClient creates a new proxy client.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		headers: make(map[string]string),
	}
}

// SetHeader sets a default header to be sent with all requests.
func (c *Client) SetHeader(key, value string) {
	c.headers[key] = value
}

// Request represents an API request.
type Request struct {
	Method      string
	Path        string
	PathParams  map[string]string
	QueryParams map[string]string
	Body        map[string]any
}

// Response represents an API response.
type Response struct {
	StatusCode int
	Data       any
	Error      string
}

// Do makes an HTTP request to the upstream API.
func (c *Client) Do(ctx context.Context, req Request) (*Response, error) {
	// Build URL path with path parameters substituted
	path := req.Path
	for key, value := range req.PathParams {
		path = replacePathParam(path, key, value)
	}

	// Build full URL
	fullURL, err := url.Parse(c.baseURL + "/" + path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters
	query := fullURL.Query()
	for key, value := range req.QueryParams {
		query.Set(key, value)
	}
	fullURL.RawQuery = query.Encode()

	// Prepare request body
	var bodyReader io.Reader
	if req.Body != nil && len(req.Body) > 0 {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, fullURL.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	httpReq.Header.Set("Accept", "application/json")
	for key, value := range c.headers {
		httpReq.Header.Set(key, value)
	}

	// Set content type for body requests
	if req.Body != nil && len(req.Body) > 0 {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	slog.Debug("proxy request",
		"method", req.Method,
		"url", fullURL.String(),
	)

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	slog.Debug("proxy response",
		"status", resp.StatusCode,
		"body_size", len(bodyBytes),
	)

	// Try to parse JSON response
	var data any
	if len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, &data); err != nil {
			// If not JSON, treat as raw string
			data = string(bodyBytes)
		}
	}

	// Check for error status codes
	if resp.StatusCode >= 400 {
		errorMsg := fmt.Sprintf("API returned HTTP %d", resp.StatusCode)
		if errMap, ok := data.(map[string]any); ok {
			if errStr, ok := errMap["error"].(string); ok {
				errorMsg += ": " + errStr
			} else if msgStr, ok := errMap["message"].(string); ok {
				errorMsg += ": " + msgStr
			}
		}
		return &Response{
			StatusCode: resp.StatusCode,
			Data:       data,
			Error:      errorMsg,
		}, nil
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Data:       data,
	}, nil
}

// replacePathParam replaces a path parameter like {id} with its value.
func replacePathParam(path, key, value string) string {
	return url.PathEscape(value)
}
