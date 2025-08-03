package pythainlp_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/tassa-yoniso-manasi-karoto/go-pythainlp"
)

func TestIntegration(t *testing.T) {
	// Skip if not explicitly enabled
	if os.Getenv("PYTHAINLP_TEST") != "1" {
		t.Skip("Integration tests disabled. Set PYTHAINLP_TEST=1 to run")
	}

	// Enable debug logging if requested
	if os.Getenv("PYTHAINLP_DEBUG") == "1" {
		pythainlp.EnableDebugLogging()
	}

	ctx := context.Background()

	// Create manager
	manager, err := pythainlp.NewManager(ctx,
		pythainlp.WithQueryTimeout(30*time.Second))
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Initialize
	t.Log("Initializing PyThaiNLP container...")
	if err := manager.Init(ctx); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}
	defer manager.Close()

	// Test cases
	testText := "สวัสดีครับ ผมชื่อโกโก้"

	t.Run("Tokenize", func(t *testing.T) {
		result, err := manager.Tokenize(ctx, testText)
		if err != nil {
			t.Fatalf("Tokenize failed: %v", err)
		}

		t.Logf("Tokens: %v", result.Raw)
		t.Logf("Engine: %s, Processing time: %.2fms", result.Engine, result.ProcessingTime)

		if len(result.Raw) == 0 {
			t.Error("Expected tokens, got none")
		}
	})

	t.Run("TokenizeWithEngine", func(t *testing.T) {
		engines := []string{pythainlp.EngineNewMM, pythainlp.EngineLongest}
		
		for _, engine := range engines {
			result, err := manager.TokenizeWithEngine(ctx, testText, engine)
			if err != nil {
				t.Logf("Engine %s not available: %v", engine, err)
				continue
			}

			t.Logf("Engine %s: %v", engine, result.Raw)
		}
	})

	t.Run("Romanize", func(t *testing.T) {
		result, err := manager.Romanize(ctx, testText)
		if err != nil {
			t.Fatalf("Romanize failed: %v", err)
		}

		t.Logf("Romanized: %s", result.Text)
		t.Logf("Engine: %s, Processing time: %.2fms", result.Engine, result.ProcessingTime)

		if result.Text == "" {
			t.Error("Expected romanized text, got empty")
		}
	})

	t.Run("RomanizeWithTokens", func(t *testing.T) {
		opts := pythainlp.RomanizeOptions{
			Engine:        pythainlp.EngineRoyin,
			TokenizeFirst: true,
		}

		result, err := manager.RomanizeWithOptions(ctx, testText, opts)
		if err != nil {
			t.Fatalf("RomanizeWithOptions failed: %v", err)
		}

		t.Logf("Tokens: %v", result.Tokens)
		t.Logf("Romanized parts: %v", result.RomanizedParts)
		t.Logf("Full romanized: %s", result.Text)
	})

	t.Run("Transliterate", func(t *testing.T) {
		// Note: Some engines require additional dependencies
		engines := []string{pythainlp.EngineICUTrans}
		
		for _, engine := range engines {
			result, err := manager.TransliterateWithEngine(ctx, "สวัสดี", engine)
			if err != nil {
				t.Logf("Engine %s not available: %v", engine, err)
				continue
			}

			t.Logf("Engine %s: %s", engine, result.Phonetic)
		}
	})

	t.Run("AnalyzeText", func(t *testing.T) {
		result, err := manager.AnalyzeText(ctx, testText)
		if err != nil {
			t.Fatalf("AnalyzeText failed: %v", err)
		}

		t.Logf("Raw tokens: %v", result.RawTokens)
		t.Logf("Romanized: %s", result.Romanized)
		t.Logf("Processing time: %.2fms", result.ProcessingTime)

		if len(result.Tokens) > 0 {
			t.Log("Token details:")
			for i, token := range result.Tokens {
				t.Logf("  [%d] Surface: %s, Romanized: %s, IsLexical: %v",
					i, token.Surface, token.Romanization, token.IsLexical)
			}
		}
	})

	t.Run("GetVersion", func(t *testing.T) {
		version, err := manager.GetVersion(ctx)
		if err != nil {
			t.Fatalf("GetVersion failed: %v", err)
		}

		t.Logf("PyThaiNLP version: %s", version)
	})

	t.Run("GetSupportedEngines", func(t *testing.T) {
		engines, err := manager.GetSupportedEngines(ctx)
		if err != nil {
			t.Fatalf("GetSupportedEngines failed: %v", err)
		}

		for category, engineList := range engines {
			t.Logf("%s engines: %v", category, engineList)
		}
	})
}

func TestPackageLevelFunctions(t *testing.T) {
	// Skip if not explicitly enabled
	if os.Getenv("PYTHAINLP_TEST") != "1" {
		t.Skip("Integration tests disabled. Set PYTHAINLP_TEST=1 to run")
	}

	// Initialize default instance
	if err := pythainlp.Init(); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}
	defer pythainlp.Close()

	// Simple test with package-level functions
	testText := "ภาษาไทย"

	t.Run("PackageTokenize", func(t *testing.T) {
		result, err := pythainlp.Tokenize(testText)
		if err != nil {
			t.Fatalf("Tokenize failed: %v", err)
		}

		t.Logf("Tokens: %v", result.Raw)
	})

	t.Run("PackageRomanize", func(t *testing.T) {
		result, err := pythainlp.Romanize(testText)
		if err != nil {
			t.Fatalf("Romanize failed: %v", err)
		}

		t.Logf("Romanized: %s", result.Text)
	})

	t.Run("PackageAnalyze", func(t *testing.T) {
		result, err := pythainlp.TokenizeAndRomanize(testText)
		if err != nil {
			t.Fatalf("TokenizeAndRomanize failed: %v", err)
		}

		t.Logf("Tokens: %v", result.RawTokens)
		t.Logf("Romanized: %s", result.Romanized)
	})
}