package pythainlp

import (
	"context"
	"fmt"
)

// AnalyzeText performs combined analysis with tokenization and romanization
func (pm *PyThaiNLPManager) AnalyzeText(ctx context.Context, text string) (*AnalyzeResult, error) {
	opts := AnalyzeOptions{
		Features: []string{"tokenize", "romanize"},
	}
	return pm.AnalyzeWithOptions(ctx, text, opts)
}

// AnalyzeWithOptions performs combined analysis with specified options
func (pm *PyThaiNLPManager) AnalyzeWithOptions(ctx context.Context, text string, opts AnalyzeOptions) (*AnalyzeResult, error) {
	if !pm.IsReady() {
		return nil, fmt.Errorf("service not ready")
	}

	// Prepare request
	req := &AnalyzeRequest{
		Text:                text,
		Features:            opts.Features,
		TokenizeEngine:      opts.TokenizeEngine,
		RomanizeEngine:      opts.RomanizeEngine,
		TransliterateEngine: opts.TransliterateEngine,
	}

	// Set default features if not specified
	if len(req.Features) == 0 {
		req.Features = []string{"tokenize", "romanize"}
	}

	// Make API call
	resp, err := pm.client.Analyze(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	// Extract processing time
	var processingTime float64
	if v, ok := resp.Metadata["processing_time_ms"].(float64); ok {
		processingTime = v
	}

	// Build result
	result := &AnalyzeResult{
		RawTokens:      resp.Data.Tokens,
		Romanized:      resp.Data.Romanized,
		RomanizedParts: resp.Data.RomanizedTokens,
		Phonetic:       resp.Data.Phonetic,
		Features:       req.Features,
		ProcessingTime: processingTime,
	}

	// Create Token objects
	if len(resp.Data.Tokens) > 0 {
		result.Tokens = make([]Token, len(resp.Data.Tokens))
		for i, token := range resp.Data.Tokens {
			t := Token{
				Surface:   token,
				IsLexical: isThaiText(token),
			}
			
			// Add romanization if available
			if len(resp.Data.RomanizedTokens) > i {
				t.Romanization = resp.Data.RomanizedTokens[i]
			}
			
			result.Tokens[i] = t
		}
	}

	return result, nil
}

// TokenizeAndRomanize is a convenience method for common use case
func (pm *PyThaiNLPManager) TokenizeAndRomanize(ctx context.Context, text string) (*AnalyzeResult, error) {
	return pm.AnalyzeText(ctx, text)
}

// GetSupportedEngines returns the list of supported engines for each operation
func (pm *PyThaiNLPManager) GetSupportedEngines(ctx context.Context) (map[string][]string, error) {
	if !pm.IsReady() {
		return nil, fmt.Errorf("service not ready")
	}

	health, err := pm.client.Health(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get engine info: %w", err)
	}

	return health.Engines, nil
}

// GetVersion returns the PyThaiNLP version
func (pm *PyThaiNLPManager) GetVersion(ctx context.Context) (string, error) {
	if !pm.IsReady() {
		return "", fmt.Errorf("service not ready")
	}

	health, err := pm.client.Health(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}

	return health.Version, nil
}

// Package-level convenience functions

// AnalyzeText performs combined analysis with tokenization and romanization
func AnalyzeText(text string) (*AnalyzeResult, error) {
	ctx := context.Background()
	mgr, err := getOrCreateDefaultManager(ctx)
	if err != nil {
		return nil, err
	}
	return mgr.AnalyzeText(ctx, text)
}

// AnalyzeWithOptions performs combined analysis with specified options
func AnalyzeWithOptions(text string, opts AnalyzeOptions) (*AnalyzeResult, error) {
	ctx := context.Background()
	mgr, err := getOrCreateDefaultManager(ctx)
	if err != nil {
		return nil, err
	}
	return mgr.AnalyzeWithOptions(ctx, text, opts)
}

// TokenizeAndRomanize is a convenience function for common use case
func TokenizeAndRomanize(text string) (*AnalyzeResult, error) {
	return AnalyzeText(text)
}

// GetSupportedEngines returns the list of supported engines
func GetSupportedEngines() (map[string][]string, error) {
	ctx := context.Background()
	mgr, err := getOrCreateDefaultManager(ctx)
	if err != nil {
		return nil, err
	}
	return mgr.GetSupportedEngines(ctx)
}

// GetVersion returns the PyThaiNLP version
func GetVersion() (string, error) {
	ctx := context.Background()
	mgr, err := getOrCreateDefaultManager(ctx)
	if err != nil {
		return "", err
	}
	return mgr.GetVersion(ctx)
}

// Utility functions for working with results

// JoinTokens joins tokens into a single string
func JoinTokens(tokens []string) string {
	result := ""
	for i, token := range tokens {
		if i > 0 && !isThaiText(token) && !isThaiText(tokens[i-1]) {
			result += " "
		}
		result += token
	}
	return result
}

// ExtractSurfaces extracts just the surface text from Token objects
func ExtractSurfaces(tokens []Token) []string {
	surfaces := make([]string, len(tokens))
	for i, token := range tokens {
		surfaces[i] = token.Surface
	}
	return surfaces
}