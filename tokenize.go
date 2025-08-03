package pythainlp

import (
	"context"
	"fmt"
)

// Tokenize performs word tokenization using the default engine (newmm)
func (pm *PyThaiNLPManager) Tokenize(ctx context.Context, text string) (*TokenizeResult, error) {
	return pm.TokenizeWithEngine(ctx, text, EngineNewMM)
}

// TokenizeWithEngine performs word tokenization with a specified engine
func (pm *PyThaiNLPManager) TokenizeWithEngine(ctx context.Context, text string, engine string) (*TokenizeResult, error) {
	opts := TokenizeOptions{
		Engine: engine,
	}
	return pm.TokenizeWithOptions(ctx, text, opts)
}

// TokenizeWithOptions performs word tokenization with full options
func (pm *PyThaiNLPManager) TokenizeWithOptions(ctx context.Context, text string, opts TokenizeOptions) (*TokenizeResult, error) {
	if !pm.IsReady() {
		return nil, fmt.Errorf("service not ready")
	}

	// Prepare request
	req := &TokenizeRequest{
		Text:    text,
		Engine:  opts.Engine,
		Options: opts.Extra,
	}

	// Set default engine if not specified
	if req.Engine == "" {
		req.Engine = EngineNewMM
	}

	// Make API call
	resp, err := pm.client.Tokenize(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("tokenization failed: %w", err)
	}

	// Extract processing time
	var processingTime float64
	if v, ok := resp.Metadata["processing_time_ms"].(float64); ok {
		processingTime = v
	}

	// Build result
	result := &TokenizeResult{
		Raw:            resp.Tokens,
		Engine:         req.Engine,
		ProcessingTime: processingTime,
	}

	// Create Token objects with just the surface text for now
	// Future versions can add more linguistic information
	result.Tokens = make([]Token, len(resp.Tokens))
	for i, token := range resp.Tokens {
		result.Tokens[i] = Token{
			Surface:   token,
			IsLexical: isThaiText(token),
		}
	}

	return result, nil
}

// Package-level functions for backward compatibility

// Tokenize performs word tokenization using the default engine
func Tokenize(text string) (*TokenizeResult, error) {
	ctx := context.Background()
	mgr, err := getOrCreateDefaultManager(ctx)
	if err != nil {
		return nil, err
	}
	return mgr.Tokenize(ctx, text)
}

// TokenizeWithEngine performs word tokenization with a specified engine
func TokenizeWithEngine(text string, engine string) (*TokenizeResult, error) {
	ctx := context.Background()
	mgr, err := getOrCreateDefaultManager(ctx)
	if err != nil {
		return nil, err
	}
	return mgr.TokenizeWithEngine(ctx, text, engine)
}

// TokenizeWithOptions performs word tokenization with full options
func TokenizeWithOptions(text string, opts TokenizeOptions) (*TokenizeResult, error) {
	ctx := context.Background()
	mgr, err := getOrCreateDefaultManager(ctx)
	if err != nil {
		return nil, err
	}
	return mgr.TokenizeWithOptions(ctx, text, opts)
}

// Helper functions

// isThaiText checks if a token contains Thai characters
func isThaiText(text string) bool {
	for _, r := range text {
		if r >= 0x0E00 && r <= 0x0E7F {
			return true
		}
	}
	return false
}