package pythainlp

import (
	"context"
	"fmt"
)

// SyllableTokenize performs syllable tokenization using the default engine (han_solo)
func (pm *PyThaiNLPManager) SyllableTokenize(ctx context.Context, text string) (*SyllableTokenizeResult, error) {
	return pm.SyllableTokenizeWithEngine(ctx, text, EngineSyllableHanSolo)
}

// SyllableTokenizeWithEngine performs syllable tokenization with a specified engine
func (pm *PyThaiNLPManager) SyllableTokenizeWithEngine(ctx context.Context, text string, engine string) (*SyllableTokenizeResult, error) {
	opts := SyllableTokenizeOptions{
		Engine: engine,
	}
	return pm.SyllableTokenizeWithOptions(ctx, text, opts)
}

// SyllableTokenizeWithOptions performs syllable tokenization with full options
func (pm *PyThaiNLPManager) SyllableTokenizeWithOptions(ctx context.Context, text string, opts SyllableTokenizeOptions) (*SyllableTokenizeResult, error) {
	if !pm.IsReady() {
		return nil, fmt.Errorf("service not ready")
	}

	// Prepare request
	req := &SyllableTokenizeRequest{
		Text:           text,
		Engine:         opts.Engine,
		KeepWhitespace: opts.KeepWhitespace,
	}

	// Set default engine if not specified
	if req.Engine == "" {
		req.Engine = EngineSyllableHanSolo
	}

	// Make API call
	resp, err := pm.client.SyllableTokenize(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("syllable tokenization failed: %w", err)
	}

	// Extract processing time
	var processingTime float64
	if v, ok := resp.Metadata["processing_time_ms"].(float64); ok {
		processingTime = v
	}

	// Build result
	result := &SyllableTokenizeResult{
		Syllables:      resp.Syllables,
		Engine:         req.Engine,
		ProcessingTime: processingTime,
	}

	return result, nil
}

// Package-level functions for backward compatibility

// SyllableTokenize performs syllable tokenization using the default engine
func SyllableTokenize(text string) (*SyllableTokenizeResult, error) {
	ctx := context.Background()
	mgr, err := getOrCreateDefaultManager(ctx)
	if err != nil {
		return nil, err
	}
	return mgr.SyllableTokenize(ctx, text)
}

// SyllableTokenizeWithEngine performs syllable tokenization with a specified engine
func SyllableTokenizeWithEngine(text string, engine string) (*SyllableTokenizeResult, error) {
	ctx := context.Background()
	mgr, err := getOrCreateDefaultManager(ctx)
	if err != nil {
		return nil, err
	}
	return mgr.SyllableTokenizeWithEngine(ctx, text, engine)
}

// SyllableTokenizeWithOptions performs syllable tokenization with full options
func SyllableTokenizeWithOptions(text string, opts SyllableTokenizeOptions) (*SyllableTokenizeResult, error) {
	ctx := context.Background()
	mgr, err := getOrCreateDefaultManager(ctx)
	if err != nil {
		return nil, err
	}
	return mgr.SyllableTokenizeWithOptions(ctx, text, opts)
}