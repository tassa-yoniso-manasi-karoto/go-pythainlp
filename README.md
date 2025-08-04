# go-pythainlp

Basic Go bindings for PyThaiNLP using Docker containers and HTTP API.

## Features

- **Tokenization** - Word segmentation with multiple engines
- **Romanization** - Convert Thai text to Latin alphabet
- **Transliteration** - Phonetic conversion (IPA, tltk, etc.)
- **High Performance** - Persistent Python service eliminates startup overhead
- **Docker-based** - No Python installation required on host
- **Lightweight Mode** - By default use a ~170MB image without obsolete ML tokenizers or BERT sentiment (96% size reduction)

<!-- CLAUDE:
Romanization engines that need PyTorch:
- thai2rom - Deep learning based (requires PyTorch)
- thaig2p - IPA transliteration (requires PyTorch)

Romanization engines that DON'T need PyTorch:
- royin (default) - Rule-based, Royal Institute standard
- tltk - Rule-based Thai Language Toolkit
- lookup - Dictionary-based
- thai2rom_onnx - Uses ONNX runtime instead of PyTorch

Other PyTorch users (all unrelated to your needs):
- attacut/deepcut - Tokenizers you don't want
- wangchanberta - BERT models for various NLP tasks
- chat/generate modules - Text generation
- spell checking, parsing, etc.

For tokenization, the best engines don't need PyTorch:
- newmm - Dictionary-based (your preferred choice)
- longest - Dictionary-based
- nlpo3 - Rust-based
- icu - ICU library based

-->

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

## Lightweight Mode

> [!IMPORTANT]
> **go-pythainlp defaults to lightweight mode** to optimize the end-user experience. This mode downloads only 170MB (vs 3.9GB) and builds in ~4 minutes, while still providing the best-performing tokenization engine (`newmm` with F1 score of 0.802, outperforming neural engines at 0.775). Perfect for applications like subtitle processing where users need quality results without lengthy installation times.

By default, go-pythainlp uses lightweight mode which excludes neural network dependencies, reducing the Docker image size from ~3.9GB to ~170MB. This mode includes the best-performing dictionary-based engines like `newmm` and `nlpo3`.

### Available in Lightweight Mode:
- **Tokenizers**: newmm, longest, icu, nercut, tltk, nlpo3
- **Romanizers**: royin, tltk, lookup
- **Transliterators**: icu, iso_11940, tltk_ipa, tltk_g2p

### Excluded in Lightweight Mode:
- **Tokenizers**: attacut, deepcut, oskut, sefr_cut (neural networks)
- **Romanizers**: thai2rom (requires PyTorch)
- **Transliterators**: thaig2p, ipa (require PyTorch/epitran)

### Using Full Mode

To use the full PyThaiNLP with all neural network models:

```go
// Option 1: Set package variable before Init
pythainlp.UseLightweightMode = false
pythainlp.Init()

// Option 2: Use manager with option
manager, err := pythainlp.NewManager(ctx,
    pythainlp.WithLightweightMode(false))
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