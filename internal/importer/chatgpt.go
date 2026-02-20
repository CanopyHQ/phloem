// Package importer provides tools to import AI history from various sources
package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CanopyHQ/phloem/internal/memory"
)

// ChatGPTConversation represents a ChatGPT export conversation
type ChatGPTConversation struct {
	Title       string                 `json:"title"`
	CreateTime  float64                `json:"create_time"`
	UpdateTime  float64                `json:"update_time"`
	Mapping     map[string]ChatGPTNode `json:"mapping"`
	CurrentNode string                 `json:"current_node,omitempty"`
}

// ChatGPTNode represents a node in the conversation tree
type ChatGPTNode struct {
	ID       string          `json:"id"`
	Message  *ChatGPTMessage `json:"message,omitempty"`
	Parent   *string         `json:"parent,omitempty"`
	Children []string        `json:"children,omitempty"`
}

// ChatGPTMessage represents a message in ChatGPT format
type ChatGPTMessage struct {
	ID         string         `json:"id"`
	Author     ChatGPTAuthor  `json:"author"`
	CreateTime *float64       `json:"create_time,omitempty"`
	Content    ChatGPTContent `json:"content"`
	Status     string         `json:"status,omitempty"`
}

// ChatGPTAuthor represents the message author
type ChatGPTAuthor struct {
	Role string `json:"role"`
	Name string `json:"name,omitempty"`
}

// ChatGPTContent represents message content
type ChatGPTContent struct {
	ContentType string   `json:"content_type"`
	Parts       []string `json:"parts,omitempty"`
}

// ChatGPTExport represents the full export file
type ChatGPTExport []ChatGPTConversation

// ImportResult tracks import statistics
type ImportResult struct {
	ConversationsProcessed int
	MemoriesCreated        int
	Errors                 []string
	Duration               time.Duration
}

// ChatGPTImporter imports ChatGPT conversation history
type ChatGPTImporter struct {
	store *memory.Store
}

// NewChatGPTImporter creates a new ChatGPT importer
func NewChatGPTImporter(store *memory.Store) *ChatGPTImporter {
	return &ChatGPTImporter{store: store}
}

// ImportFromFile imports conversations from a ChatGPT export file
func (i *ChatGPTImporter) ImportFromFile(ctx context.Context, filePath string) (*ImportResult, error) {
	start := time.Now()
	result := &ImportResult{}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON
	var export ChatGPTExport
	if err := json.Unmarshal(data, &export); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Process each conversation
	for _, conv := range export {
		result.ConversationsProcessed++

		// Extract meaningful content from conversation
		memories := i.extractMemories(conv)

		for _, mem := range memories {
			_, err := i.store.Remember(ctx, mem.Content, mem.Tags, mem.Context)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("conversation %s: %v", conv.Title, err))
				continue
			}
			result.MemoriesCreated++
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// MemoryCandidate represents a potential memory to create
type MemoryCandidate struct {
	Content string
	Tags    []string
	Context string
}

// extractMemories extracts memorable content from a conversation
func (i *ChatGPTImporter) extractMemories(conv ChatGPTConversation) []MemoryCandidate {
	var memories []MemoryCandidate

	// Get messages in order by traversing the tree
	messages := i.flattenConversation(conv)

	// Look for patterns worth remembering
	var currentTopic string
	var userMessage string

	for _, msg := range messages {
		if msg.Message == nil || msg.Message.Content.ContentType != "text" {
			continue
		}

		content := strings.Join(msg.Message.Content.Parts, "\n")
		if content == "" {
			continue
		}

		role := msg.Message.Author.Role

		if role == "user" {
			userMessage = content
			// Extract topic from user message
			currentTopic = extractTopic(content)
		} else if role == "assistant" && userMessage != "" {
			// Create memory from Q&A pair if it looks useful
			if isWorthRemembering(userMessage, content) {
				mem := MemoryCandidate{
					Content: fmt.Sprintf("Q: %s\n\nA: %s",
						truncate(userMessage, 500),
						truncate(content, 2000)),
					Tags:    inferTags(userMessage, content),
					Context: fmt.Sprintf("From ChatGPT conversation: %s", conv.Title),
				}

				if currentTopic != "" {
					mem.Tags = append(mem.Tags, currentTopic)
				}
				mem.Tags = append(mem.Tags, "imported", "chatgpt")

				memories = append(memories, mem)
			}
			userMessage = "" // Reset for next pair
		}
	}

	return memories
}

// flattenConversation converts the tree structure to a linear list
func (i *ChatGPTImporter) flattenConversation(conv ChatGPTConversation) []ChatGPTNode {
	var result []ChatGPTNode

	// Find root nodes (no parent)
	var roots []string
	for id, node := range conv.Mapping {
		if node.Parent == nil || *node.Parent == "" {
			roots = append(roots, id)
		}
	}

	// BFS traversal
	var traverse func(id string)
	traverse = func(id string) {
		if node, ok := conv.Mapping[id]; ok {
			result = append(result, node)
			for _, childID := range node.Children {
				traverse(childID)
			}
		}
	}

	for _, root := range roots {
		traverse(root)
	}

	return result
}

// isWorthRemembering determines if a Q&A pair is worth saving
func isWorthRemembering(question, answer string) bool {
	// Skip very short exchanges
	if len(question) < 20 || len(answer) < 100 {
		return false
	}

	// Skip generic greetings
	lowerQ := strings.ToLower(question)
	skipPhrases := []string{"hello", "hi there", "hey", "thanks", "thank you", "bye", "goodbye"}
	for _, phrase := range skipPhrases {
		if strings.HasPrefix(lowerQ, phrase) {
			return false
		}
	}

	// Skip very long answers (probably not memorable)
	if len(answer) > 10000 {
		return false
	}

	return true
}

// extractTopic tries to identify the main topic from a question
func extractTopic(question string) string {
	lower := strings.ToLower(question)

	// Common topic indicators
	topics := map[string][]string{
		"code":     {"code", "function", "implement", "bug", "error", "programming"},
		"writing":  {"write", "essay", "article", "blog", "story"},
		"explain":  {"explain", "what is", "how does", "why"},
		"math":     {"calculate", "math", "equation", "formula"},
		"research": {"research", "study", "paper", "source"},
		"career":   {"job", "career", "interview", "resume"},
		"learning": {"learn", "study", "course", "tutorial"},
	}

	for topic, keywords := range topics {
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				return topic
			}
		}
	}

	return ""
}

// inferTags extracts relevant tags from content
func inferTags(question, answer string) []string {
	var tags []string
	combined := strings.ToLower(question + " " + answer)

	// Programming languages
	languages := []string{"python", "javascript", "typescript", "go", "golang", "rust", "java", "c++", "ruby", "php", "swift", "kotlin"}
	for _, lang := range languages {
		if strings.Contains(combined, lang) {
			tags = append(tags, lang)
		}
	}

	// Frameworks
	frameworks := []string{"react", "vue", "angular", "django", "flask", "express", "nextjs", "rails"}
	for _, fw := range frameworks {
		if strings.Contains(combined, fw) {
			tags = append(tags, fw)
		}
	}

	// Topics
	topicKeywords := map[string][]string{
		"api":      {"api", "rest", "graphql", "endpoint"},
		"database": {"database", "sql", "postgres", "mongodb"},
		"ai":       {"machine learning", "neural", "gpt", "llm", "ai"},
		"devops":   {"docker", "kubernetes", "ci/cd", "deploy"},
		"security": {"security", "auth", "encryption", "jwt"},
	}

	for tag, keywords := range topicKeywords {
		for _, kw := range keywords {
			if strings.Contains(combined, kw) {
				tags = append(tags, tag)
				break
			}
		}
	}

	return tags
}

// truncate shortens a string to maxLen
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// ImportFromDirectory imports all JSON files from a directory
func (i *ChatGPTImporter) ImportFromDirectory(ctx context.Context, dirPath string) (*ImportResult, error) {
	combined := &ImportResult{}
	start := time.Now()

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".json") {
			result, err := i.ImportFromFile(ctx, path)
			if err != nil {
				combined.Errors = append(combined.Errors, fmt.Sprintf("%s: %v", path, err))
				return nil // Continue with other files
			}

			combined.ConversationsProcessed += result.ConversationsProcessed
			combined.MemoriesCreated += result.MemoriesCreated
			combined.Errors = append(combined.Errors, result.Errors...)
		}

		return nil
	})

	combined.Duration = time.Since(start)
	return combined, err
}
