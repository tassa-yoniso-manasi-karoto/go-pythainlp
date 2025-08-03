# go-pythainlp

Go bindings for PyThaiNLP using Docker containers and HTTP API.

## Features

- **Tokenization** - Word segmentation with multiple engines (newmm, attacut, deepcut, etc.)
- **Romanization** - Convert Thai text to Latin alphabet (royin, thai2rom, etc.)
- **Transliteration** - Phonetic conversion (IPA, thaig2p, etc.)
- **High Performance** - Persistent Python service eliminates startup overhead
- **Docker-based** - No Python installation required on host

## Installation

```bash
go get github.com/tassa-yoniso-manasi-karoto/go-pythainlp
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    pythainlp "github.com/tassa-yoniso-manasi-karoto/go-pythainlp"
)

func main() {
    // Initialize with default settings
    if err := pythainlp.Init(); err != nil {
        log.Fatal(err)
    }
    defer pythainlp.Close()
    
    // Tokenize Thai text
    result, err := pythainlp.Tokenize("สวัสดีครับ")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Tokens: %v\n", result.Raw)
    // Output: Tokens: [สวัสดี ครับ]
    
    // Romanize Thai text
    roman, err := pythainlp.Romanize("ภาษาไทย")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Romanized: %s\n", roman.Text)
    // Output: Romanized: phasa thai
}
```

## Advanced Usage

### Custom Manager

```go
ctx := context.Background()

// Create custom manager
manager, err := pythainlp.NewManager(ctx,
    pythainlp.WithProjectName("my-project"),
    pythainlp.WithQueryTimeout(30*time.Second))
if err != nil {
    log.Fatal(err)
}

// Initialize containers
if err := manager.Init(ctx); err != nil {
    log.Fatal(err)
}
defer manager.Close()

// Use different engines
result, err := manager.TokenizeWithEngine(ctx, "ภาษาไทย", pythainlp.EngineAttaCut)
```

### Combined Analysis

```go
// Tokenize and romanize in one call
result, err := pythainlp.AnalyzeText("สวัสดีครับ ผมชื่อโกโก้")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Tokens: %v\n", result.RawTokens)
fmt.Printf("Romanized: %s\n", result.Romanized)

// Access detailed token information
for _, token := range result.Tokens {
    fmt.Printf("Token: %s, Romanized: %s\n", 
        token.Surface, token.Romanization)
}
```

## Available Engines

### Tokenization Engines
- `newmm` (default) - Dictionary-based with TCC
- `longest` - Dictionary-based, longest matching
- `attacut` - Deep learning based
- `deepcut` - Deep learning based
- `nlpo3` - Rust implementation (fast)
- Others: `icu`, `nercut`, `oskut`, `sefr_cut`, `tltk`

### Romanization Engines
- `royin` (default) - Royal Institute standard
- `thai2rom` - Deep learning based
- `tltk` - Thai Language Toolkit
- `lookup` - Dictionary lookup

### Transliteration Engines
- `thaig2p` (default) - Thai grapheme-to-phoneme
- `icu` - ICU transliteration
- `ipa` - International Phonetic Alphabet
- Others: `tltk_g2p`, `iso_11940`, `tltk_ipa`

## Requirements

- Docker Desktop (Windows/Mac) or Docker Engine (Linux)
- Go 1.19 or later

## Testing

```bash
# Run integration tests
PYTHAINLP_TEST=1 go test -v ./...

# Run with debug logging
PYTHAINLP_TEST=1 PYTHAINLP_DEBUG=1 go test -v ./...
```

## Debug Logging

To enable debug logging:

```go
import pythainlp "github.com/tassa-yoniso-manasi-karoto/go-pythainlp"

// Enable debug logs
pythainlp.EnableDebugLogging()

// Now all operations will log detailed information
pythainlp.Init()
```

## Performance

The persistent service architecture provides:
- Sub-millisecond latency for simple operations
- One-time initialization cost (10-30 seconds)
- Handles 1000+ requests/second
- Shared model instances reduce memory usage

## License

GPL 3