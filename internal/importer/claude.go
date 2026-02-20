package importer

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CanopyHQ/phloem/internal/memory"
)

// ClaudeConversation represents a Claude export conversation
type ClaudeConversation struct {
	UUID         string          `json:"uuid"`
	Name         string          `json:"name"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	ChatMessages []ClaudeMessage `json:"chat_messages"`
}

// ClaudeMessage represents a message in Claude format
type ClaudeMessage struct {
	UUID      string    `json:"uuid"`
	Text      string    `json:"text"`
	Sender    string    `json:"sender"` // "human" or "assistant"
	CreatedAt time.Time `json:"created_at"`
}

// ClaudeImporter imports Claude conversation history
type ClaudeImporter struct {
	store *memory.Store
}

// NewClaudeImporter creates a new Claude importer
func NewClaudeImporter(store *memory.Store) *ClaudeImporter {
	return &ClaudeImporter{store: store}
}

// ImportFromFile imports conversations from a Claude export file
func (i *ClaudeImporter) ImportFromFile(ctx context.Context, filePath string) (*ImportResult, error) {
	start := time.Now()
	result := &ImportResult{}

	// Read file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Claude exports can be JSONL (one conversation per line) or JSON array
	ext := strings.ToLower(filepath.Ext(filePath))

	var conversations []ClaudeConversation

	if ext == ".jsonl" {
		// JSONL format
		scanner := bufio.NewScanner(file)
		// Increase buffer size for long conversations
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 10*1024*1024) // 10MB max

		for scanner.Scan() {
			var conv ClaudeConversation
			if err := json.Unmarshal(scanner.Bytes(), &conv); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("line parse error: %v", err))
				continue
			}
			conversations = append(conversations, conv)
		}

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("scanner error: %w", err)
		}
	} else {
		// JSON array format
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&conversations); err != nil {
			// Try single conversation
			file.Seek(0, 0)
			var single ClaudeConversation
			if err := json.NewDecoder(file).Decode(&single); err != nil {
				return nil, fmt.Errorf("failed to parse JSON: %w", err)
			}
			conversations = []ClaudeConversation{single}
		}
	}

	// Process each conversation
	for _, conv := range conversations {
		result.ConversationsProcessed++

		memories := i.extractMemories(conv)

		for _, mem := range memories {
			_, err := i.store.Remember(ctx, mem.Content, mem.Tags, mem.Context)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("conversation %s: %v", conv.Name, err))
				continue
			}
			result.MemoriesCreated++
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// extractMemories extracts memorable content from a Claude conversation
func (i *ClaudeImporter) extractMemories(conv ClaudeConversation) []MemoryCandidate {
	var memories []MemoryCandidate

	var userMessage string
	var currentTopic string

	for _, msg := range conv.ChatMessages {
		content := strings.TrimSpace(msg.Text)
		if content == "" {
			continue
		}

		if msg.Sender == "human" {
			userMessage = content
			currentTopic = extractTopic(content)
		} else if msg.Sender == "assistant" && userMessage != "" {
			// Create memory from Q&A pair if it looks useful
			if isWorthRemembering(userMessage, content) {
				mem := MemoryCandidate{
					Content: fmt.Sprintf("Q: %s\n\nA: %s",
						truncate(userMessage, 500),
						truncate(content, 2000)),
					Tags:    inferTags(userMessage, content),
					Context: fmt.Sprintf("From Claude conversation: %s", conv.Name),
				}

				if currentTopic != "" {
					mem.Tags = append(mem.Tags, currentTopic)
				}
				mem.Tags = append(mem.Tags, "imported", "claude")

				memories = append(memories, mem)
			}
			userMessage = "" // Reset for next pair
		}
	}

	return memories
}

// ImportFromDirectory imports all JSON/JSONL files from a directory
func (i *ClaudeImporter) ImportFromDirectory(ctx context.Context, dirPath string) (*ImportResult, error) {
	combined := &ImportResult{}
	start := time.Now()

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		lower := strings.ToLower(path)
		if !info.IsDir() && (strings.HasSuffix(lower, ".json") || strings.HasSuffix(lower, ".jsonl")) {
			result, err := i.ImportFromFile(ctx, path)
			if err != nil {
				combined.Errors = append(combined.Errors, fmt.Sprintf("%s: %v", path, err))
				return nil
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
