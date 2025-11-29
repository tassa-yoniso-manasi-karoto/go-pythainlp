package pythainlp

import (
	"context"
	"fmt"
)

// Romanize performs romanization using the default engine (royin)
func (pm *PyThaiNLPManager) Romanize(ctx context.Context, text string) (*RomanizeResult, error) {
	return pm.RomanizeWithEngine(ctx, text, EngineRoyin)
}

// RomanizeWithEngine performs romanization with a specified engine
func (pm *PyThaiNLPManager) RomanizeWithEngine(ctx context.Context, text string, engine string) (*RomanizeResult, error) {
	opts := RomanizeOptions{
		Engine: engine,
	}
	return pm.RomanizeWithOptions(ctx, text, opts)
}

// RomanizeWithOptions performs romanization with full options
func (pm *PyThaiNLPManager) RomanizeWithOptions(ctx context.Context, text string, opts RomanizeOptions) (*RomanizeResult, error) {
	if !pm.IsReady() {
		return nil, fmt.Errorf("service not ready")
	}

	// Prepare request
	req := &RomanizeRequest{
		Text:     text,
		Engine:   opts.Engine,
		Tokenize: opts.TokenizeFirst,
	}

	// Set default engine if not specified
	if req.Engine == "" {
		req.Engine = EngineRoyin
	}

	// Make API call
	resp, err := pm.client.Romanize(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("romanization failed: %w", err)
	}

	// Extract processing time
	var processingTime float64
	if v, ok := resp.Metadata["processing_time_ms"].(float64); ok {
		processingTime = v
	}

	// Build result
	result := &RomanizeResult{
		Text:           resp.Romanized,
		Tokens:         resp.Tokens,
		RomanizedParts: resp.RomanizedTokens,
		Engine:         req.Engine,
		ProcessingTime: processingTime,
	}

	return result, nil
}

// Transliterate performs transliteration (phonetic conversion) using the default engine (thaig2p)
func (pm *PyThaiNLPManager) Transliterate(ctx context.Context, text string) (*TransliterateResult, error) {
	return pm.TransliterateWithEngine(ctx, text, EngineThaig2p)
}

// TransliterateWithEngine performs transliteration with a specified engine
func (pm *PyThaiNLPManager) TransliterateWithEngine(ctx context.Context, text string, engine string) (*TransliterateResult, error) {
	opts := TransliterateOptions{
		Engine: engine,
	}
	return pm.TransliterateWithOptions(ctx, text, opts)
}

// TransliterateWithOptions performs transliteration with full options
func (pm *PyThaiNLPManager) TransliterateWithOptions(ctx context.Context, text string, opts TransliterateOptions) (*TransliterateResult, error) {
	if !pm.IsReady() {
		return nil, fmt.Errorf("service not ready")
	}

	// Prepare request
	req := &TransliterateRequest{
		Text:   text,
		Engine: opts.Engine,
	}

	// Set default engine if not specified
	if req.Engine == "" {
		req.Engine = EngineThaig2p
	}

	// Make API call
	resp, err := pm.client.Transliterate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("transliteration failed: %w", err)
	}

	// Extract processing time
	var processingTime float64
	if v, ok := resp.Metadata["processing_time_ms"].(float64); ok {
		processingTime = v
	}

	// Build result
	result := &TransliterateResult{
		Phonetic:       resp.Phonetic,
		Engine:         req.Engine,
		ProcessingTime: processingTime,
	}

	return result, nil
}

// Pronunciate is an alias for Transliterate, following PyThaiNLP naming
func (pm *PyThaiNLPManager) Pronunciate(ctx context.Context, text string) (*TransliterateResult, error) {
	return pm.Transliterate(ctx, text)
}

// Package-level functions for backward compatibility

// Romanize performs romanization using the default engine
func Romanize(text string) (*RomanizeResult, error) {
	ctx := context.Background()
	mgr, err := getOrCreateDefaultManager(ctx)
	if err != nil {
		return nil, err
	}
	return mgr.Romanize(ctx, text)
}

// RomanizeWithEngine performs romanization with a specified engine
func RomanizeWithEngine(text string, engine string) (*RomanizeResult, error) {
	ctx := context.Background()
	mgr, err := getOrCreateDefaultManager(ctx)
	if err != nil {
		return nil, err
	}
	return mgr.RomanizeWithEngine(ctx, text, engine)
}

// RomanizeWithOptions performs romanization with full options
func RomanizeWithOptions(text string, opts RomanizeOptions) (*RomanizeResult, error) {
	ctx := context.Background()
	mgr, err := getOrCreateDefaultManager(ctx)
	if err != nil {
		return nil, err
	}
	return mgr.RomanizeWithOptions(ctx, text, opts)
}

// Transliterate performs transliteration using the default engine
func Transliterate(text string) (*TransliterateResult, error) {
	ctx := context.Background()
	mgr, err := getOrCreateDefaultManager(ctx)
	if err != nil {
		return nil, err
	}
	return mgr.Transliterate(ctx, text)
}

// TransliterateWithEngine performs transliteration with a specified engine
func TransliterateWithEngine(text string, engine string) (*TransliterateResult, error) {
	ctx := context.Background()
	mgr, err := getOrCreateDefaultManager(ctx)
	if err != nil {
		return nil, err
	}
	return mgr.TransliterateWithEngine(ctx, text, engine)
}

// TransliterateWithOptions performs transliteration with full options
func TransliterateWithOptions(text string, opts TransliterateOptions) (*TransliterateResult, error) {
	ctx := context.Background()
	mgr, err := getOrCreateDefaultManager(ctx)
	if err != nil {
		return nil, err
	}
	return mgr.TransliterateWithOptions(ctx, text, opts)
}

// Pronunciate is an alias for Transliterate
func Pronunciate(text string) (*TransliterateResult, error) {
	return Transliterate(text)
}