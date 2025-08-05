package pythainlp

// Token represents a single token with linguistic information
// This is a subset of tha.Tkn from translitkit, focused on essential fields
type Token struct {
	// Core fields
	Surface      string `json:"surface"`      // The token text
	Romanization string `json:"romanization"` // Romanized form
	IPA          string `json:"ipa"`          // IPA phonetic representation
	
	// Linguistic properties
	POS       string `json:"pos,omitempty"`       // Part of speech tag
	IsLexical bool   `json:"is_lexical"`          // Whether it's Thai text or punctuation/foreign
	
	// Additional metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"` // Engine-specific data
}

// TokenizeResult contains the results of tokenization
type TokenizeResult struct {
	Tokens []Token  // Structured tokens with linguistic info
	Raw    []string // Simple tokenized strings
	
	// Metadata
	Engine         string  `json:"engine"`
	ProcessingTime float64 `json:"processing_time_ms"`
}

// RomanizeResult contains the results of romanization
type RomanizeResult struct {
	Text           string   // Full romanized text
	Tokens         []string // Original tokens (if tokenized first)
	RomanizedParts []string // Per-token romanization
	
	// Metadata
	Engine         string  `json:"engine"`
	ProcessingTime float64 `json:"processing_time_ms"`
}

// TransliterateResult contains the results of transliteration (phonetic)
type TransliterateResult struct {
	Phonetic string // IPA or other phonetic representation
	
	// Metadata
	Engine         string  `json:"engine"`
	ProcessingTime float64 `json:"processing_time_ms"`
}

// SyllableTokenizeResult contains the results of syllable tokenization
type SyllableTokenizeResult struct {
	Syllables []string // Syllable segments
	
	// Metadata
	Engine         string  `json:"engine"`
	ProcessingTime float64 `json:"processing_time_ms"`
}

// AnalyzeResult contains combined analysis results
type AnalyzeResult struct {
	Tokens         []Token  // Structured tokens
	RawTokens      []string // Simple token strings
	Romanized      string   // Full romanized text
	RomanizedParts []string // Per-token romanization
	Phonetic       string   // IPA representation
	Syllables      []string // Syllable segments
	
	// Metadata
	Features       []string `json:"features"`
	ProcessingTime float64  `json:"processing_time_ms"`
}

// Engine constants for tokenization
const (
	EngineNewMM    = "newmm"    // Default, dictionary-based with TCC
	EngineLongest  = "longest"  // Dictionary-based, longest matching
	EngineICU      = "icu"      // ICU-based tokenizer
	EngineAttaCut  = "attacut"  // Deep learning based
	EngineDeepCut  = "deepcut"  // Deep learning based
	EngineNerCut   = "nercut"   // NER-aware tokenizer
	EngineNLPO3    = "nlpo3"    // Rust-based, fast
	EngineOSKut    = "oskut"    // Out-of-domain stacked cut
	EngineSefrCut  = "sefr_cut" // Stacked ensemble
	EngineTLTK     = "tltk"     // Maximum collocation
)

// Engine constants for romanization
const (
	EngineRoyin    = "royin"    // Default, Royal Institute standard
	EngineThai2Rom = "thai2rom" // Deep learning based
	EngineTLTKRom  = "tltk"     // TLTK romanization
	EngineLookup   = "lookup"   // Dictionary lookup
)

// Engine constants for transliteration
const (
	EngineThaig2p   = "thaig2p"   // Default, Thai grapheme-to-phoneme
	EngineICUTrans  = "icu"       // ICU transliteration
	EngineIPA       = "ipa"       // Epitran IPA
	EngineTLTKG2P   = "tltk_g2p"  // TLTK grapheme-to-phoneme
	EngineISO11940  = "iso_11940" // ISO 11940 standard
	EngineTLTKIPA   = "tltk_ipa"  // TLTK IPA
	EngineThaig2pV2 = "thaig2p_v2" // Version 2 of thaig2p
)

// Engine constants for syllable tokenization
const (
	EngineSyllableDict    = "dict"     // Dictionary-based syllable tokenization
	EngineSyllableHanSolo = "han_solo" // Default, CRF syllable segmenter for social media
	EngineSyllableSSG     = "ssg"      // CRF syllable segmenter
	EngineSyllableTLTK    = "tltk"     // Thai Language Toolkit syllable tokenizer
)

// Options for various operations
type TokenizeOptions struct {
	Engine         string                 // Tokenization engine to use
	CustomDict     []string               // Custom dictionary entries
	KeepWhitespace bool                   // Whether to keep whitespace tokens
	JoinBrokenNum  bool                   // Join broken numbers
	Extra          map[string]interface{} // Engine-specific options
}

type RomanizeOptions struct {
	Engine          string // Romanization engine to use
	TokenizeFirst   bool   // Whether to tokenize before romanizing
	FallbackEngine  string // Fallback for lookup engine
}

type TransliterateOptions struct {
	Engine string // Transliteration engine to use
}

type SyllableTokenizeOptions struct {
	Engine         string // Syllable tokenization engine to use
	KeepWhitespace bool   // Whether to keep whitespace tokens
}

type AnalyzeOptions struct {
	Features            []string // Features to extract: tokenize, romanize, transliterate, syllable
	TokenizeEngine      string   // Engine for tokenization
	RomanizeEngine      string   // Engine for romanization
	TransliterateEngine string   // Engine for transliteration
	SyllableEngine      string   // Engine for syllable tokenization
}

// Error types
type PyThaiNLPError struct {
	Code    string
	Message string
	Details map[string]interface{}
}

func (e PyThaiNLPError) Error() string {
	return e.Message
}