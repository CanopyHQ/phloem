package memory

import (
	"math"
	"os"
	"testing"
)

func TestLocalEmbedder_Embed(t *testing.T) {
	embedder := NewLocalEmbedder()

	tests := []struct {
		name string
		text string
	}{
		{"simple", "hello world"},
		{"code", "function foo() { return bar; }"},
		{"question", "how do I fix this bug?"},
		{"empty", ""},
		{"long", "This is a longer piece of text that contains multiple sentences. It should generate a meaningful embedding that captures the semantic content."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			embedding, err := embedder.Embed(tt.text)
			if err != nil {
				t.Fatalf("Embed() error = %v", err)
			}

			if len(embedding) != embedder.Dimensions() {
				t.Errorf("Embed() returned %d dimensions, want %d", len(embedding), embedder.Dimensions())
			}

			// Check normalization (should be unit vector or zero)
			var norm float32
			for _, v := range embedding {
				norm += v * v
			}
			norm = float32(math.Sqrt(float64(norm)))

			if tt.text != "" && (norm < 0.99 || norm > 1.01) {
				t.Errorf("Embed() not normalized, norm = %f", norm)
			}
		})
	}
}

func TestLocalEmbedder_Similarity(t *testing.T) {
	embedder := NewLocalEmbedder()

	// Similar texts should have higher similarity than dissimilar texts
	tests := []struct {
		name  string
		text1 string
		text2 string
		text3 string // Should be less similar to text1 than text2
	}{
		{
			name:  "code_similarity",
			text1: "fix the bug in the login function",
			text2: "debug the error in the authentication method",
			text3: "what time is the meeting tomorrow",
		},
		{
			name:  "question_similarity",
			text1: "how do I deploy to production?",
			text2: "what's the process for deploying?",
			text3: "the weather is nice today",
		},
		{
			name:  "topic_similarity",
			text1: "implement user authentication with OAuth",
			text2: "add login functionality using OAuth2",
			text3: "buy groceries from the store",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emb1, _ := embedder.Embed(tt.text1)
			emb2, _ := embedder.Embed(tt.text2)
			emb3, _ := embedder.Embed(tt.text3)

			sim12 := testCosineSimilarity(emb1, emb2)
			sim13 := testCosineSimilarity(emb1, emb3)

			if sim12 <= sim13 {
				t.Errorf("Expected similarity(%q, %q) > similarity(%q, %q), got %f <= %f",
					tt.text1, tt.text2, tt.text1, tt.text3, sim12, sim13)
			}
		})
	}
}

func TestLocalEmbedder_EmbedBatch(t *testing.T) {
	embedder := NewLocalEmbedder()

	texts := []string{
		"first text",
		"second text",
		"third text",
	}

	embeddings, err := embedder.EmbedBatch(texts)
	if err != nil {
		t.Fatalf("EmbedBatch() error = %v", err)
	}

	if len(embeddings) != len(texts) {
		t.Errorf("EmbedBatch() returned %d embeddings, want %d", len(embeddings), len(texts))
	}

	// Each embedding should match individual Embed() call
	for i, text := range texts {
		single, _ := embedder.Embed(text)
		if !testVectorsEqual(embeddings[i], single) {
			t.Errorf("EmbedBatch()[%d] != Embed(%q)", i, text)
		}
	}
}

func TestLocalEmbedder_SemanticCategories(t *testing.T) {
	embedder := NewLocalEmbedder()

	// Code-related text should have different embedding than time-related
	codeText := "implement the function and fix the bug in the API endpoint"
	timeText := "schedule the meeting for tomorrow morning at 9am"

	codeEmb, _ := embedder.Embed(codeText)
	timeEmb, _ := embedder.Embed(timeText)

	// They should be different (low similarity)
	sim := testCosineSimilarity(codeEmb, timeEmb)
	if sim > 0.8 {
		t.Errorf("Code and time texts too similar: %f", sim)
	}
}

func TestLocalEmbedder_Deterministic(t *testing.T) {
	embedder := NewLocalEmbedder()
	text := "this is a test of deterministic embeddings"

	emb1, _ := embedder.Embed(text)
	emb2, _ := embedder.Embed(text)

	if !testVectorsEqual(emb1, emb2) {
		t.Error("Embed() not deterministic - same text produced different embeddings")
	}
}

// Helper functions

func testCosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dot, normA, normB float32
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

func testVectorsEqual(a, b []float32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if math.Abs(float64(a[i]-b[i])) > 1e-6 {
			return false
		}
	}
	return true
}

func TestNewCambiumEmbedder_NotImplemented(t *testing.T) {
	emb, err := NewCambiumEmbedder()
	if err == nil {
		t.Fatal("expected error (Cambium embeddings not yet implemented)")
	}
	if emb != nil {
		t.Error("expected nil embedder")
	}
}

func TestNewOpenAIEmbedder_NoAPIKey(t *testing.T) {
	// Ensure no key is set for this test
	old := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer func() { os.Setenv("OPENAI_API_KEY", old) }()

	emb, err := NewOpenAIEmbedder()
	if err == nil {
		t.Fatal("expected error when OPENAI_API_KEY not set")
	}
	if emb != nil {
		t.Error("expected nil embedder")
	}
}

func TestNewGeminiEmbedder_NoAPIKey(t *testing.T) {
	old := os.Getenv("GEMINI_API_KEY")
	os.Unsetenv("GEMINI_API_KEY")
	defer func() { os.Setenv("GEMINI_API_KEY", old) }()

	emb, err := NewGeminiEmbedder()
	if err == nil {
		t.Fatal("expected error when GEMINI_API_KEY not set")
	}
	if emb != nil {
		t.Error("expected nil embedder")
	}
}

func TestGetEmbedder_AirGapped(t *testing.T) {
	os.Setenv("PHLOEM_AIR_GAPPED", "1")
	defer os.Unsetenv("PHLOEM_AIR_GAPPED")

	emb := GetEmbedder()
	if emb == nil {
		t.Fatal("GetEmbedder returned nil")
	}
	if emb.Dimensions() != 512 {
		t.Errorf("air-gapped should use local embedder (512 dims), got %d", emb.Dimensions())
	}
}
