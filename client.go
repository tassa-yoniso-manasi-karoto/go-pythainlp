package pythainlp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles HTTP communication with the Python service
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new HTTP client for the PyThaiNLP service
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// ServiceError represents an error returned by the Python service
type ServiceError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e ServiceError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ServiceResponse is the common response structure from all endpoints
type ServiceResponse struct {
	Data     json.RawMessage        `json:"data"`
	Metadata map[string]interface{} `json:"metadata"`
	Error    *ServiceError          `json:"error"`
}

// doRequest performs an HTTP request and handles the response
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*ServiceResponse, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var serviceResp ServiceResponse
	if err := json.Unmarshal(respBody, &serviceResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if serviceResp.Error != nil {
		return nil, serviceResp.Error
	}

	return &serviceResp, nil
}

// Health checks the service health status
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	// Health endpoint returns plain JSON, not wrapped
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var health HealthResponse
	if err := json.Unmarshal(body, &health); err != nil {
		return nil, fmt.Errorf("failed to parse health response: %w", err)
	}

	return &health, nil
}

// Tokenize performs word tokenization
func (c *Client) Tokenize(ctx context.Context, req *TokenizeRequest) (*TokenizeResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/tokenize", req)
	if err != nil {
		return nil, err
	}

	var data struct {
		Tokens []string `json:"tokens"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse tokenize response: %w", err)
	}

	return &TokenizeResponse{
		Tokens:   data.Tokens,
		Metadata: resp.Metadata,
	}, nil
}

// Romanize performs romanization
func (c *Client) Romanize(ctx context.Context, req *RomanizeRequest) (*RomanizeResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/romanize", req)
	if err != nil {
		return nil, err
	}

	var data struct {
		Romanized       string   `json:"romanized"`
		Tokens          []string `json:"tokens,omitempty"`
		RomanizedTokens []string `json:"romanized_tokens,omitempty"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse romanize response: %w", err)
	}

	return &RomanizeResponse{
		Romanized:       data.Romanized,
		Tokens:          data.Tokens,
		RomanizedTokens: data.RomanizedTokens,
		Metadata:        resp.Metadata,
	}, nil
}

// Transliterate performs transliteration (phonetic conversion)
func (c *Client) Transliterate(ctx context.Context, req *TransliterateRequest) (*TransliterateResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/transliterate", req)
	if err != nil {
		return nil, err
	}

	var data struct {
		Phonetic string `json:"phonetic"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse transliterate response: %w", err)
	}

	return &TransliterateResponse{
		Phonetic: data.Phonetic,
		Metadata: resp.Metadata,
	}, nil
}

// Analyze performs combined analysis
func (c *Client) Analyze(ctx context.Context, req *AnalyzeRequest) (*AnalyzeResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/analyze", req)
	if err != nil {
		return nil, err
	}

	var data AnalyzeData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse analyze response: %w", err)
	}

	return &AnalyzeResponse{
		Data:     data,
		Metadata: resp.Metadata,
	}, nil
}

// Request types

// TokenizeRequest represents a tokenization request
type TokenizeRequest struct {
	Text    string                 `json:"text"`
	Engine  string                 `json:"engine,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// RomanizeRequest represents a romanization request
type RomanizeRequest struct {
	Text     string `json:"text"`
	Engine   string `json:"engine,omitempty"`
	Tokenize bool   `json:"tokenize,omitempty"`
}

// TransliterateRequest represents a transliteration request
type TransliterateRequest struct {
	Text   string `json:"text"`
	Engine string `json:"engine,omitempty"`
}

// AnalyzeRequest represents a combined analysis request
type AnalyzeRequest struct {
	Text                string   `json:"text"`
	Features            []string `json:"features"`
	TokenizeEngine      string   `json:"tokenize_engine,omitempty"`
	RomanizeEngine      string   `json:"romanize_engine,omitempty"`
	TransliterateEngine string   `json:"transliterate_engine,omitempty"`
}

// Response types

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string              `json:"status"`
	Version string              `json:"version"`
	Engines map[string][]string `json:"engines"`
}

// TokenizeResponse represents a tokenization response
type TokenizeResponse struct {
	Tokens   []string               `json:"tokens"`
	Metadata map[string]interface{} `json:"metadata"`
}

// RomanizeResponse represents a romanization response
type RomanizeResponse struct {
	Romanized       string                 `json:"romanized"`
	Tokens          []string               `json:"tokens,omitempty"`
	RomanizedTokens []string               `json:"romanized_tokens,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// TransliterateResponse represents a transliteration response
type TransliterateResponse struct {
	Phonetic string                 `json:"phonetic"`
	Metadata map[string]interface{} `json:"metadata"`
}

// AnalyzeData contains the results of combined analysis
type AnalyzeData struct {
	Tokens          []string `json:"tokens,omitempty"`
	Romanized       string   `json:"romanized,omitempty"`
	RomanizedTokens []string `json:"romanized_tokens,omitempty"`
	Phonetic        string   `json:"phonetic,omitempty"`
}

// AnalyzeResponse represents a combined analysis response
type AnalyzeResponse struct {
	Data     AnalyzeData            `json:"data"`
	Metadata map[string]interface{} `json:"metadata"`
}