// Package memory provides embedding generation for semantic search
package memory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Embedder generates vector embeddings for text
type Embedder interface {
	Embed(text string) ([]float32, error)
	EmbedBatch(texts []string) ([][]float32, error)
	Dimensions() int
}

// FallbackEmbedder wraps a primary embedder and falls back to local on errors (e.g. expired API keys)
type FallbackEmbedder struct {
	primary  Embedder
	fallback Embedder
	failed   bool // sticky: once primary fails, stay on fallback for the session
}

func NewFallbackEmbedder(primary Embedder) *FallbackEmbedder {
	return &FallbackEmbedder{
		primary:  primary,
		fallback: NewLocalEmbedder(),
	}
}

func (f *FallbackEmbedder) Embed(text string) ([]float32, error) {
	if f.failed {
		return f.fallback.Embed(text)
	}
	result, err := f.primary.Embed(text)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Primary embedder failed (%v), falling back to local\n", err)
		f.failed = true
		return f.fallback.Embed(text)
	}
	return result, nil
}

func (f *FallbackEmbedder) EmbedBatch(texts []string) ([][]float32, error) {
	if f.failed {
		return f.fallback.EmbedBatch(texts)
	}
	result, err := f.primary.EmbedBatch(texts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Primary embedder failed (%v), falling back to local\n", err)
		f.failed = true
		return f.fallback.EmbedBatch(texts)
	}
	return result, nil
}

func (f *FallbackEmbedder) Dimensions() int {
	if f.failed {
		return f.fallback.Dimensions()
	}
	return f.primary.Dimensions()
}

// CambiumEmbedder uses Cambium proxy for embeddings (routes to best available provider)
type CambiumEmbedder struct {
	baseURL    string
	model      string
	dimensions int
	client     *http.Client
}

// NewCambiumEmbedder creates an embedder using Cambium proxy
// NOTE: Cambium doesn't currently support embeddings endpoint, so this is disabled
func NewCambiumEmbedder() (*CambiumEmbedder, error) {
	// Cambium doesn't have embeddings endpoint yet - skip for now
	// TODO: Enable when Cambium adds /v1/embeddings support
	return nil, fmt.Errorf("Cambium embeddings not yet implemented")

	/*
		// Check if Cambium is running - try common ports
		baseURL := os.Getenv("CAMBIUM_URL")

		client := &http.Client{Timeout: 2 * time.Second}

		// Try configured URL first, then common ports
		urlsToTry := []string{}
		if baseURL != "" {
			urlsToTry = append(urlsToTry, baseURL)
		}
		urlsToTry = append(urlsToTry, "http://localhost:8080", "https://cambium.canopyhq.io")

		for _, url := range urlsToTry {
			resp, err := client.Get(url + "/health")
			if err == nil && resp.StatusCode == 200 {
				resp.Body.Close()
				return &CambiumEmbedder{
					baseURL:    url,
					model:      "text-embedding-3-small", // Cambium will route appropriately
					dimensions: 1536,
					client: &http.Client{
						Timeout: 30 * time.Second,
					},
				}, nil
			}
			if resp != nil {
				resp.Body.Close()
			}
		}

		return nil, fmt.Errorf("Cambium not available on any port")
	*/
}

// OpenAIEmbedder uses OpenAI's embedding API directly (fallback if Cambium not running)
type OpenAIEmbedder struct {
	apiKey     string
	model      string
	dimensions int
	client     *http.Client
}

// NewOpenAIEmbedder creates an embedder using OpenAI's API directly
func NewOpenAIEmbedder() (*OpenAIEmbedder, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not set")
	}

	return &OpenAIEmbedder{
		apiKey:     apiKey,
		model:      "text-embedding-3-small",
		dimensions: 1536,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// GeminiEmbedder uses Google's Gemini embedding API
type GeminiEmbedder struct {
	apiKey     string
	model      string
	dimensions int
	client     *http.Client
}

// NewGeminiEmbedder creates an embedder using Gemini's API
func NewGeminiEmbedder() (*GeminiEmbedder, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not set")
	}

	return &GeminiEmbedder{
		apiKey:     apiKey,
		model:      "text-embedding-004", // or gemini-embedding-001
		dimensions: 768,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Embed generates an embedding via Gemini
func (e *GeminiEmbedder) Embed(text string) ([]float32, error) {
	embeddings, err := e.EmbedBatch([]string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return embeddings[0], nil
}

// EmbedBatch generates embeddings via Gemini
func (e *GeminiEmbedder) EmbedBatch(texts []string) ([][]float32, error) {
	// Gemini uses a different API format
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent?key=%s", e.model, e.apiKey)

	embeddings := make([][]float32, len(texts))

	for i, text := range texts {
		reqBody := map[string]interface{}{
			"content": map[string]interface{}{
				"parts": []map[string]string{
					{"text": text},
				},
			},
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := e.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Gemini API error %d: %s", resp.StatusCode, string(body))
		}

		var result struct {
			Embedding struct {
				Values []float32 `json:"values"`
			} `json:"embedding"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		embeddings[i] = result.Embedding.Values
	}

	return embeddings, nil
}

// Dimensions returns the embedding dimension size
func (e *GeminiEmbedder) Dimensions() int {
	return e.dimensions
}

// Embed generates an embedding for a single text via Cambium
func (e *CambiumEmbedder) Embed(text string) ([]float32, error) {
	embeddings, err := e.EmbedBatch([]string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return embeddings[0], nil
}

// EmbedBatch generates embeddings for multiple texts via Cambium
func (e *CambiumEmbedder) EmbedBatch(texts []string) ([][]float32, error) {
	return callEmbeddingAPI(e.client, e.baseURL+"/v1/embeddings", "", e.model, texts)
}

// Dimensions returns the embedding dimension size
func (e *CambiumEmbedder) Dimensions() int {
	return e.dimensions
}

// Embed generates an embedding for a single text via OpenAI directly
func (e *OpenAIEmbedder) Embed(text string) ([]float32, error) {
	embeddings, err := e.EmbedBatch([]string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return embeddings[0], nil
}

// EmbedBatch generates embeddings for multiple texts via OpenAI directly
func (e *OpenAIEmbedder) EmbedBatch(texts []string) ([][]float32, error) {
	return callEmbeddingAPI(e.client, "https://api.openai.com/v1/embeddings", e.apiKey, e.model, texts)
}

// Dimensions returns the embedding dimension size
func (e *OpenAIEmbedder) Dimensions() int {
	return e.dimensions
}

// callEmbeddingAPI is shared logic for calling OpenAI-compatible embedding APIs
func callEmbeddingAPI(client *http.Client, url, apiKey, model string, texts []string) ([][]float32, error) {
	reqBody := map[string]interface{}{
		"model": model,
		"input": texts,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Sort by index to maintain order
	embeddings := make([][]float32, len(texts))
	for _, d := range result.Data {
		if d.Index < len(embeddings) {
			embeddings[d.Index] = d.Embedding
		}
	}

	return embeddings, nil
}

// LocalEmbedder uses enhanced on-device embeddings for offline/free operation
// Combines multiple techniques for better semantic similarity than basic TF-IDF:
// 1. N-gram hashing (captures word order and phrases)
// 2. Character-level features (handles typos and variations)
// 3. Positional weighting (emphasizes beginning/end of text)
// 4. Semantic category boosting (common semantic markers)
type LocalEmbedder struct {
	dimensions int
	ngramSizes []int
	stopwords  map[string]bool
}

// NewLocalEmbedder creates a local embedder with enhanced features
func NewLocalEmbedder() *LocalEmbedder {
	return &LocalEmbedder{
		dimensions: 512,            // Larger for better quality
		ngramSizes: []int{1, 2, 3}, // Unigrams, bigrams, trigrams
		stopwords:  buildStopwords(),
	}
}

// buildStopwords returns common English stopwords
func buildStopwords() map[string]bool {
	words := []string{
		"the", "a", "an", "and", "or", "but", "in", "on", "at", "to", "for",
		"of", "with", "by", "from", "as", "is", "was", "are", "were", "been",
		"be", "have", "has", "had", "do", "does", "did", "will", "would", "could",
		"should", "may", "might", "must", "shall", "can", "need", "dare", "ought",
		"used", "it", "its", "this", "that", "these", "those", "i", "you", "he",
		"she", "we", "they", "what", "which", "who", "whom", "whose", "where",
		"when", "why", "how", "all", "each", "every", "both", "few", "more",
		"most", "other", "some", "such", "no", "nor", "not", "only", "own",
		"same", "so", "than", "too", "very", "just", "also", "now", "here",
	}
	m := make(map[string]bool)
	for _, w := range words {
		m[w] = true
	}
	return m
}

// Semantic categories for boosting related terms
var semanticCategories = map[string][]string{
	"code":     {"function", "class", "method", "variable", "code", "programming", "bug", "fix", "error", "debug", "test", "api", "endpoint", "database", "query", "server", "client", "request", "response", "json", "http", "git", "commit", "branch", "merge", "deploy"},
	"time":     {"today", "yesterday", "tomorrow", "week", "month", "year", "morning", "afternoon", "evening", "night", "now", "later", "soon", "recently", "always", "never", "sometimes", "often", "daily", "weekly", "monthly"},
	"action":   {"create", "build", "make", "add", "remove", "delete", "update", "change", "modify", "fix", "implement", "design", "plan", "review", "test", "deploy", "launch", "start", "stop", "run", "execute"},
	"people":   {"user", "customer", "client", "team", "member", "developer", "engineer", "designer", "manager", "admin", "owner", "person", "people", "everyone", "someone", "anyone"},
	"status":   {"done", "complete", "finished", "pending", "waiting", "blocked", "progress", "started", "failed", "success", "error", "working", "broken", "fixed", "ready", "todo"},
	"priority": {"important", "urgent", "critical", "high", "low", "medium", "priority", "asap", "immediately", "soon", "later", "eventually"},
}

// Embed generates an enhanced local embedding
func (e *LocalEmbedder) Embed(text string) ([]float32, error) {
	return e.generateEnhancedEmbedding(text), nil
}

// EmbedBatch generates embeddings for multiple texts
func (e *LocalEmbedder) EmbedBatch(texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		embeddings[i] = e.generateEnhancedEmbedding(text)
	}
	return embeddings, nil
}

// Dimensions returns the embedding dimension size
func (e *LocalEmbedder) Dimensions() int {
	return e.dimensions
}

// generateEnhancedEmbedding creates a multi-feature embedding
func (e *LocalEmbedder) generateEnhancedEmbedding(text string) []float32 {
	embedding := make([]float32, e.dimensions)

	// Normalize text
	text = strings.ToLower(text)
	words := tokenize(text)

	if len(words) == 0 {
		return embedding
	}

	// 1. N-gram features (60% of dimensions)
	ngramDims := int(float64(e.dimensions) * 0.6)
	e.addNgramFeatures(embedding[:ngramDims], words)

	// 2. Character-level features (20% of dimensions)
	charStart := ngramDims
	charDims := int(float64(e.dimensions) * 0.2)
	e.addCharFeatures(embedding[charStart:charStart+charDims], text)

	// 3. Semantic category features (10% of dimensions)
	semStart := charStart + charDims
	semDims := int(float64(e.dimensions) * 0.1)
	e.addSemanticFeatures(embedding[semStart:semStart+semDims], words)

	// 4. Structural features (10% of dimensions)
	structStart := semStart + semDims
	e.addStructuralFeatures(embedding[structStart:], text, words)

	// Normalize the full embedding
	normalize(embedding)

	return embedding
}

// tokenize splits text into words, handling punctuation
func tokenize(text string) []string {
	// Replace common punctuation with spaces
	for _, p := range []string{".", ",", "!", "?", ";", ":", "'", "\"", "(", ")", "[", "]", "{", "}", "\n", "\t"} {
		text = strings.ReplaceAll(text, p, " ")
	}

	words := strings.Fields(text)
	result := make([]string, 0, len(words))

	for _, word := range words {
		word = strings.TrimSpace(word)
		if len(word) > 1 { // Skip single characters
			result = append(result, word)
		}
	}

	return result
}

// addNgramFeatures adds n-gram based features
func (e *LocalEmbedder) addNgramFeatures(embedding []float32, words []string) {
	dims := len(embedding)

	for _, n := range e.ngramSizes {
		weight := 1.0 / float32(n) // Smaller n-grams get more weight

		for i := 0; i <= len(words)-n; i++ {
			ngram := strings.Join(words[i:i+n], " ")

			// Skip if all stopwords
			allStop := true
			for j := i; j < i+n; j++ {
				if !e.stopwords[words[j]] {
					allStop = false
					break
				}
			}
			if allStop && n == 1 {
				continue
			}

			// Hash to multiple positions (feature hashing)
			h1 := hashString(ngram)
			h2 := hashString(ngram + "_2")

			idx1 := h1 % dims
			idx2 := h2 % dims

			// Positional weighting - words at start/end matter more
			posWeight := float32(1.0)
			if i < 3 || i >= len(words)-3 {
				posWeight = 1.5
			}

			// TF component
			tfWeight := float32(1.0 + math.Log(float64(1+countOccurrences(words, ngram, n))))

			embedding[idx1] += weight * posWeight * tfWeight
			embedding[idx2] -= weight * posWeight * tfWeight * 0.5 // Negative for diversity
		}
	}
}

// countOccurrences counts how many times an n-gram appears
func countOccurrences(words []string, ngram string, n int) int {
	count := 0
	for i := 0; i <= len(words)-n; i++ {
		if strings.Join(words[i:i+n], " ") == ngram {
			count++
		}
	}
	return count
}

// addCharFeatures adds character-level features (handles typos, variations)
func (e *LocalEmbedder) addCharFeatures(embedding []float32, text string) {
	dims := len(embedding)

	// Character trigrams
	for i := 0; i < len(text)-2; i++ {
		trigram := text[i : i+3]
		h := hashString("char_" + trigram)
		idx := h % dims
		embedding[idx] += 0.1
	}

	// Character distribution (vowels, consonants, digits, special)
	vowels := 0
	consonants := 0
	digits := 0
	special := 0

	for _, c := range text {
		switch {
		case strings.ContainsRune("aeiou", c):
			vowels++
		case c >= 'a' && c <= 'z':
			consonants++
		case c >= '0' && c <= '9':
			digits++
		case c != ' ':
			special++
		}
	}

	total := float32(len(text))
	if total > 0 && dims >= 4 {
		embedding[0] = float32(vowels) / total
		embedding[1] = float32(consonants) / total
		embedding[2] = float32(digits) / total
		embedding[3] = float32(special) / total
	}
}

// addSemanticFeatures adds category-based semantic features
func (e *LocalEmbedder) addSemanticFeatures(embedding []float32, words []string) {
	dims := len(embedding)
	if dims == 0 {
		return
	}

	categoryScores := make(map[string]float32)

	for _, word := range words {
		for category, keywords := range semanticCategories {
			for _, kw := range keywords {
				if word == kw || strings.Contains(word, kw) {
					categoryScores[category] += 1.0
				}
			}
		}
	}

	// Map categories to embedding positions
	categories := []string{"code", "time", "action", "people", "status", "priority"}
	for i, cat := range categories {
		if i < dims {
			embedding[i] = categoryScores[cat] / float32(len(words)+1)
		}
	}
}

// addStructuralFeatures adds text structure features
func (e *LocalEmbedder) addStructuralFeatures(embedding []float32, text string, words []string) {
	dims := len(embedding)
	if dims < 8 {
		return
	}

	// Length features
	embedding[0] = float32(math.Log(float64(len(text) + 1)))
	embedding[1] = float32(math.Log(float64(len(words) + 1)))

	// Average word length
	totalLen := 0
	for _, w := range words {
		totalLen += len(w)
	}
	if len(words) > 0 {
		embedding[2] = float32(totalLen) / float32(len(words))
	}

	// Sentence count (approximate)
	sentences := strings.Count(text, ".") + strings.Count(text, "!") + strings.Count(text, "?")
	embedding[3] = float32(math.Log(float64(sentences + 1)))

	// Question indicator
	if strings.Contains(text, "?") {
		embedding[4] = 1.0
	}

	// Code indicator (backticks, common code patterns)
	if strings.Contains(text, "`") || strings.Contains(text, "()") || strings.Contains(text, "{}") {
		embedding[5] = 1.0
	}

	// List indicator
	if strings.Contains(text, "- ") || strings.Contains(text, "* ") || strings.Contains(text, "1.") {
		embedding[6] = 1.0
	}

	// Uppercase ratio (emphasis)
	upperCount := 0
	for _, c := range text {
		if c >= 'A' && c <= 'Z' {
			upperCount++
		}
	}
	if len(text) > 0 {
		embedding[7] = float32(upperCount) / float32(len(text))
	}
}

// normalize normalizes a vector to unit length
func normalize(v []float32) {
	var norm float32
	for _, x := range v {
		norm += x * x
	}
	if norm > 0 {
		norm = float32(math.Sqrt(float64(norm)))
		for i := range v {
			v[i] /= norm
		}
	}
}

// GetEmbedder returns the best available embedder based on deployment mode and license tier.
//
// EMBEDDING STRATEGY BY USE CASE:
// - Air-gapped: Local only (PHLOEM_AIR_GAPPED=1) - no API, no sync
// - Canopy Org: API embeddings (performance, price insensitive) - set PHLOEM_ORG_MODE=true
// - Admin-phloem: API embeddings (low latency, high resilience) - set PHLOEM_ADMIN_MODE=true
// - Pro tier: API embeddings (cloud or choice); fallback to local if no API keys
// - Free tier / unlicensed: Local embeddings by default (privacy + cost; no PHLOEM_EMBEDDINGS needed)
//
// Explicit override: Set PHLOEM_EMBEDDINGS=openai|gemini|cambium|local (Pro can use cloud; free defaults to local)
func GetEmbedder() Embedder {
	embedder := getEmbedderInner()
	// Wrap any API-based embedder with fallback to local on runtime errors
	// (e.g. expired API keys, network failures)
	if _, isLocal := embedder.(*LocalEmbedder); !isLocal {
		return NewFallbackEmbedder(embedder)
	}
	return embedder
}

func getEmbedderInner() Embedder {
	// 0. Air-gapped: local embedder only, no API calls
	if os.Getenv("PHLOEM_AIR_GAPPED") == "1" || os.Getenv("PHLOEM_AIR_GAPPED") == "true" {
		return NewLocalEmbedder()
	}

	// 1. Check explicit override first
	embedMode := os.Getenv("PHLOEM_EMBEDDINGS")
	if embedMode != "" {
		switch embedMode {
		case "openai":
			if os.Getenv("OPENAI_API_KEY") != "" {
				embedder, err := NewOpenAIEmbedder()
				if err == nil {
					fmt.Fprintln(os.Stderr, "🧠 Using OpenAI embeddings (explicit override)")
					return embedder
				}
				fmt.Fprintf(os.Stderr, "⚠️  OpenAI embedder failed: %v, falling back\n", err)
			} else {
				fmt.Fprintln(os.Stderr, "⚠️  PHLOEM_EMBEDDINGS=openai but OPENAI_API_KEY not set")
			}
		case "gemini":
			if os.Getenv("GEMINI_API_KEY") != "" {
				embedder, err := NewGeminiEmbedder()
				if err == nil {
					fmt.Fprintln(os.Stderr, "🧠 Using Gemini embeddings (explicit override)")
					return embedder
				}
				fmt.Fprintf(os.Stderr, "⚠️  Gemini embedder failed: %v, falling back\n", err)
			} else {
				fmt.Fprintln(os.Stderr, "⚠️  PHLOEM_EMBEDDINGS=gemini but GEMINI_API_KEY not set")
			}
		case "cambium":
			cambiumEmbedder, err := NewCambiumEmbedder()
			if err == nil {
				fmt.Fprintln(os.Stderr, "🧠 Using Cambium proxy for embeddings (explicit override)")
				return cambiumEmbedder
			}
			fmt.Fprintf(os.Stderr, "⚠️  Cambium embedder failed: %v, falling back\n", err)
		case "local":
			fmt.Fprintln(os.Stderr, "🧠 Using local embeddings (explicit override)")
			return NewLocalEmbedder()
		}
	}

	// 2. Canopy Org mode: Performance-focused, price insensitive → API embeddings
	if os.Getenv("PHLOEM_ORG_MODE") == "true" || os.Getenv("PHLOEM_ORG_MODE") == "1" {
		// Try OpenAI first (best quality)
		if os.Getenv("OPENAI_API_KEY") != "" {
			embedder, err := NewOpenAIEmbedder()
			if err == nil {
				fmt.Fprintln(os.Stderr, "🧠 Using OpenAI embeddings (Canopy org mode - performance)")
				return embedder
			}
		}
		// Fallback to Gemini
		if os.Getenv("GEMINI_API_KEY") != "" {
			embedder, err := NewGeminiEmbedder()
			if err == nil {
				fmt.Fprintln(os.Stderr, "🧠 Using Gemini embeddings (Canopy org mode - performance)")
				return embedder
			}
		}
		fmt.Fprintln(os.Stderr, "⚠️  Canopy org mode but no API keys found, falling back to local")
	}

	// 3. Admin mode: Low latency, high resilience → API embeddings
	if os.Getenv("PHLOEM_ADMIN_MODE") == "true" || os.Getenv("PHLOEM_ADMIN_MODE") == "1" {
		// Try OpenAI first (best quality)
		if os.Getenv("OPENAI_API_KEY") != "" {
			embedder, err := NewOpenAIEmbedder()
			if err == nil {
				fmt.Fprintln(os.Stderr, "🧠 Using OpenAI embeddings (admin mode - low latency)")
				return embedder
			}
		}
		// Fallback to Gemini
		if os.Getenv("GEMINI_API_KEY") != "" {
			embedder, err := NewGeminiEmbedder()
			if err == nil {
				fmt.Fprintln(os.Stderr, "🧠 Using Gemini embeddings (admin mode - low latency)")
				return embedder
			}
		}
		fmt.Fprintln(os.Stderr, "⚠️  Admin mode but no API keys found, falling back to local")
	}

	// 4. Check license tier (Pro = API embeddings for intra-device sync + AI features)
	tier := loadLicenseTier()
	if tier == "pro" {
		// Pro tier: Use API embeddings for better quality and AI memory features
		if os.Getenv("OPENAI_API_KEY") != "" {
			embedder, err := NewOpenAIEmbedder()
			if err == nil {
				fmt.Fprintln(os.Stderr, "🧠 Using OpenAI embeddings (Pro tier - intra-device sync enabled)")
				return embedder
			}
		}
		if os.Getenv("GEMINI_API_KEY") != "" {
			embedder, err := NewGeminiEmbedder()
			if err == nil {
				fmt.Fprintln(os.Stderr, "🧠 Using Gemini embeddings (Pro tier - intra-device sync enabled)")
				return embedder
			}
		}
	}

	// 5. Default: Local embeddings (privacy + cost)
	fmt.Fprintln(os.Stderr, "🧠 Using local embeddings")
	return NewLocalEmbedder()
}

// loadLicenseTier checks the license file to determine tier (avoids circular dependency)
func loadLicenseTier() string {
	dataDir := os.Getenv("PHLOEM_DATA_DIR")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "free" // Default to free
		}
		dataDir = filepath.Join(home, ".phloem")
	}

	licensePath := filepath.Join(dataDir, "license.json")
	data, err := os.ReadFile(licensePath)
	if err != nil {
		return "free" // No license = free tier
	}

	var license struct {
		Tier      string    `json:"tier"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	if err := json.Unmarshal(data, &license); err != nil {
		return "free"
	}

	// Check expiration
	if license.Tier == "pro" && !license.ExpiresAt.IsZero() {
		if time.Now().After(license.ExpiresAt) {
			return "free" // Expired = free tier
		}
	}

	return license.Tier
}
