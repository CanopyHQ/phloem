// Package mcp implements the Model Context Protocol server for Phloem
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/CanopyHQ/phloem/internal/memory"
)

// Server implements the MCP protocol over stdio
type Server struct {
	store   *memory.Store
	scanner *bufio.Scanner
}

// MemoryStats contains statistics about the memory store
type MemoryStats struct {
	TotalMemories int    `json:"total_memories"`
	DatabaseSize  string `json:"database_size"`
	LastActivity  string `json:"last_activity"`
}

// NewServer creates a new MCP server
func NewServer() (*Server, error) {
	store, err := memory.NewStore()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize memory store: %w", err)
	}
	return &Server{
		store:   store,
		scanner: bufio.NewScanner(os.Stdin),
	}, nil
}

// Start begins the MCP server loop
func (s *Server) Start() error {
	fmt.Fprintln(os.Stderr, "🧠 Phloem MCP server ready")

	for s.scanner.Scan() {
		line := s.scanner.Text()
		if line == "" {
			continue
		}

		var request JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &request); err != nil {
			s.sendError(nil, -32700, "Parse error", err.Error())
			continue
		}

		s.handleRequest(&request)
	}

	return s.scanner.Err()
}

// Stop gracefully shuts down the server
func (s *Server) Stop() {
	if s.store != nil {
		s.store.Close()
	}
}

// GetMemoryStats returns statistics about the memory store
func (s *Server) GetMemoryStats() MemoryStats {
	count, _ := s.store.Count(context.Background())
	size, _ := s.store.Size()
	lastActivity, _ := s.store.LastActivity(context.Background())

	lastActivityStr := "never"
	if !lastActivity.IsZero() {
		lastActivityStr = lastActivity.Format(time.RFC3339)
	}

	return MemoryStats{
		TotalMemories: count,
		DatabaseSize:  size,
		LastActivity:  lastActivityStr,
	}
}

// handleRequest processes a JSON-RPC request
func (s *Server) handleRequest(req *JSONRPCRequest) {
	ctx := context.Background()

	switch req.Method {
	case "initialize":
		s.handleInitialize(req)
	case "tools/list":
		s.handleToolsList(req)
	case "tools/call":
		s.handleToolCall(ctx, req)
	case "resources/list":
		s.handleResourcesList(req)
	case "resources/read":
		s.handleResourceRead(ctx, req)
	case "prompts/list":
		s.handlePromptsList(req)
	case "prompts/get":
		s.handlePromptsGet(ctx, req)
	default:
		s.sendError(req.ID, -32601, "Method not found", req.Method)
	}
}

// handleInitialize responds to the initialize request
func (s *Server) handleInitialize(req *JSONRPCRequest) {
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools":     map[string]interface{}{},
			"resources": map[string]interface{}{},
			"prompts":   map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "phloem-mcp",
			"version": "0.1.0",
		},
	}
	s.sendResult(req.ID, result)
}

// handleToolsList returns available tools
func (s *Server) handleToolsList(req *JSONRPCRequest) {
	tools := []map[string]interface{}{
		{
			"name":        "remember",
			"description": "Store a memory for later recall. Use this to save important context, decisions, code patterns, or anything the AI should remember.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"content": map[string]interface{}{
						"type":        "string",
						"description": "The content to remember",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Tags to categorize the memory (e.g., 'code', 'decision', 'pattern')",
					},
					"context": map[string]interface{}{
						"type":        "string",
						"description": "Optional context about when/where this memory applies",
					},
					"citations": map[string]interface{}{
						"type":        "array",
						"description": "Optional citations linking this memory to code/document locations",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"file_path": map[string]interface{}{
									"type":        "string",
									"description": "Path to the file",
								},
								"start_line": map[string]interface{}{
									"type":        "integer",
									"description": "Starting line number",
								},
								"end_line": map[string]interface{}{
									"type":        "integer",
									"description": "Ending line number",
								},
								"commit_sha": map[string]interface{}{
									"type":        "string",
									"description": "Optional Git commit SHA",
								},
								"content": map[string]interface{}{
									"type":        "string",
									"description": "Optional content snippet for verification",
								},
							},
							"required": []string{"file_path", "start_line", "end_line"},
						},
					},
				},
				"required": []string{"content"},
			},
		},
		{
			"name":        "recall",
			"description": "Search memories by semantic similarity. Use this to find relevant past context, decisions, or patterns.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "What you're looking for",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of memories to return (default: 5)",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Filter by tags",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			"name":        "forget",
			"description": "Delete a specific memory by ID",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "The ID of the memory to forget",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			"name":        "list_memories",
			"description": "List recent memories, optionally filtered by tags",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of memories to return (default: 10)",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Filter by tags",
					},
				},
			},
		},
		{
			"name":        "memory_stats",
			"description": "Get statistics about the memory store",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			"name":        "session_context",
			"description": "Load session context with recent memories, key decisions, and relevant background. Call this at the start of a session to restore context from previous conversations.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"hint": map[string]interface{}{
						"type":        "string",
						"description": "Optional hint about what kind of context to prioritize (e.g., 'phloem architecture', 'business setup')",
					},
				},
			},
		},
		{
			"name":        "add_citation",
			"description": "Add a citation linking a memory to a specific location in code or documents. Citations enable verification that memories are still accurate.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"memory_id": map[string]interface{}{
						"type":        "string",
						"description": "The ID of the memory to cite",
					},
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file being cited",
					},
					"start_line": map[string]interface{}{
						"type":        "integer",
						"description": "Starting line number (1-indexed)",
					},
					"end_line": map[string]interface{}{
						"type":        "integer",
						"description": "Ending line number (1-indexed)",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Snapshot of the cited content for verification",
					},
				},
				"required": []string{"memory_id", "file_path"},
			},
		},
		{
			"name":        "verify_citation",
			"description": "Verify if a citation is still valid by checking if the file content matches. Returns updated confidence score.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"citation_id": map[string]interface{}{
						"type":        "string",
						"description": "The ID of the citation to verify",
					},
				},
				"required": []string{"citation_id"},
			},
		},
		{
			"name":        "get_citations",
			"description": "Get all citations for a memory",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"memory_id": map[string]interface{}{
						"type":        "string",
						"description": "The ID of the memory",
					},
				},
				"required": []string{"memory_id"},
			},
		},
		{
			"name":        "verify_memory",
			"description": "Verify if a memory is still accurate by checking all its citations. Updates confidence scores.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"memory_id": map[string]interface{}{
						"type":        "string",
						"description": "The ID of the memory to verify",
					},
				},
				"required": []string{"memory_id"},
			},
		},
		{
			"name":        "causal_query",
			"description": "Query causal graph: 'neighbors' = memories directly linked by causal edges; 'affected' = memories that would be affected if this memory changed (transitive downstream).",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"memory_id": map[string]interface{}{
						"type":        "string",
						"description": "The ID of the memory to query",
					},
					"query_type": map[string]interface{}{
						"type":        "string",
						"description": "One of: 'neighbors' (direct causal neighbors), 'affected' (transitive descendants)",
					},
				},
				"required": []string{"memory_id", "query_type"},
			},
		},
		{
			"name":        "compose",
			"description": "Compose two recall queries: run semantic recall for query_a and query_b, merge and deduplicate results. Use when you need context that spans two topics.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query_a": map[string]interface{}{
						"type":        "string",
						"description": "First recall query",
					},
					"query_b": map[string]interface{}{
						"type":        "string",
						"description": "Second recall query",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Max results per query before merge (default 5)",
					},
				},
				"required": []string{"query_a", "query_b"},
			},
		},
		{
			"name":        "prefetch",
			"description": "Get memory suggestions to preload for the current context (e.g. file path, repo, or topic). Call when context changes to warm the session.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"context_hint": map[string]interface{}{
						"type":        "string",
						"description": "Current context (file path, repo name, or topic). If empty, returns recent important memories.",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Max suggestions (default 5)",
					},
				},
			},
		},
		{
			"name":        "prefetch_suggest",
			"description": "Suggest memories to preload given current context (e.g. open file path or last query). Use to reduce latency when the user is likely to ask about this context next.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"context": map[string]interface{}{
						"type":        "string",
						"description": "Current context (file path, topic, or last query)",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Max suggestions (default: 5)",
					},
				},
				"required": []string{"context"},
			},
		},
	}

	s.sendResult(req.ID, map[string]interface{}{"tools": tools})
}

// handleToolCall executes a tool
func (s *Server) handleToolCall(ctx context.Context, req *JSONRPCRequest) {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "Invalid params", err.Error())
		return
	}

	var result interface{}
	var err error

	switch params.Name {
	case "remember":
		result, err = s.toolRemember(ctx, params.Arguments)
	case "recall":
		result, err = s.toolRecall(ctx, params.Arguments)
	case "forget":
		result, err = s.toolForget(ctx, params.Arguments)
	case "list_memories":
		result, err = s.toolListMemories(ctx, params.Arguments)
	case "memory_stats":
		result, err = s.toolMemoryStats(ctx)
	case "session_context":
		result, err = s.toolSessionContext(ctx, params.Arguments)
	case "add_citation":
		result, err = s.toolAddCitation(ctx, params.Arguments)
	case "verify_citation":
		result, err = s.toolVerifyCitation(ctx, params.Arguments)
	case "get_citations":
		result, err = s.toolGetCitations(ctx, params.Arguments)
	case "verify_memory":
		result, err = s.toolVerifyMemory(ctx, params.Arguments)
	case "causal_query":
		result, err = s.toolCausalQuery(ctx, params.Arguments)
	case "compose":
		result, err = s.toolCompose(ctx, params.Arguments)
	case "prefetch":
		result, err = s.toolPrefetch(ctx, params.Arguments)
	case "prefetch_suggest":
		result, err = s.toolPrefetchSuggest(ctx, params.Arguments)
	default:
		s.sendError(req.ID, -32602, "Unknown tool", params.Name)
		return
	}

	if err != nil {
		s.sendResult(req.ID, map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": fmt.Sprintf("Error: %v", err)},
			},
			"isError": true,
		})
		return
	}

	// Format result as MCP content
	text, _ := json.MarshalIndent(result, "", "  ")
	s.sendResult(req.ID, map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": string(text)},
		},
	})
}

// Tool implementations

func (s *Server) toolRemember(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	content, ok := args["content"].(string)
	if !ok || content == "" {
		return nil, fmt.Errorf("content is required")
	}

	var tags []string
	if tagsRaw, ok := args["tags"].([]interface{}); ok {
		for _, t := range tagsRaw {
			if ts, ok := t.(string); ok {
				tags = append(tags, ts)
			}
		}
	}

	context := ""
	if c, ok := args["context"].(string); ok {
		context = c
	}

	mem, err := s.store.Remember(ctx, content, tags, context)
	if err != nil {
		return nil, err
	}
	memID := mem.ID

	// Add citations if provided
	var citationsAdded int
	if citationsRaw, ok := args["citations"].([]interface{}); ok {
		for _, citRaw := range citationsRaw {
			cit, ok := citRaw.(map[string]interface{})
			if !ok {
				continue
			}

			filePath, _ := cit["file_path"].(string)
			startLine, _ := cit["start_line"].(float64)
			endLine, _ := cit["end_line"].(float64)
			commitSHA, _ := cit["commit_sha"].(string)
			citContent, _ := cit["content"].(string)

			if filePath != "" {
				_, err := s.store.AddCitation(ctx, memID, filePath, int(startLine), int(endLine), commitSHA, citContent)
				if err == nil {
					citationsAdded++
				}
			}
		}
	}

	response := map[string]interface{}{
		"status":  "remembered",
		"id":      memID,
		"message": fmt.Sprintf("Memory stored with ID %s", memID),
	}
	if citationsAdded > 0 {
		response["citations_added"] = citationsAdded
		response["message"] = fmt.Sprintf("Memory stored with ID %s and %d citation(s)", memID, citationsAdded)
	}

	return response, nil
}

func (s *Server) toolRecall(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query is required")
	}

	limit := 5
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	var tags []string
	if tagsRaw, ok := args["tags"].([]interface{}); ok {
		for _, t := range tagsRaw {
			if ts, ok := t.(string); ok {
				tags = append(tags, ts)
			}
		}
	}

	var memories []*memory.Memory
	var err error
	if len(tags) > 0 {
		memories, err = s.store.Recall(ctx, query, limit, tags)
		if err != nil {
			return nil, err
		}
	} else {
		options := memory.RecallOptions{ConfidenceWeight: 0.15}
		memories, err = s.store.RecallWithRecencyBoost(ctx, query, limit, options)
		if err != nil {
			memories, err = s.store.Recall(ctx, query, limit, nil)
			if err != nil {
				return nil, err
			}
		}
	}

	results := make([]map[string]interface{}, len(memories))
	for i, mem := range memories {
		confidence := mem.Confidence
		if confidence == 0 {
			confidence, _ = s.store.GetMemoryConfidence(ctx, mem.ID)
		}

		results[i] = map[string]interface{}{
			"id":         mem.ID,
			"content":    mem.Content,
			"tags":       mem.Tags,
			"context":    mem.Context,
			"created_at": mem.CreatedAt.Format(time.RFC3339),
			"similarity": mem.Similarity,
			"confidence": confidence,
			"source":     "local",
		}
	}

	return map[string]interface{}{
		"query":    query,
		"count":    len(results),
		"memories": results,
	}, nil
}

func (s *Server) toolForget(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	id, ok := args["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("id is required")
	}

	if err := s.store.Forget(ctx, id); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status":  "forgotten",
		"id":      id,
		"message": fmt.Sprintf("Memory %s has been forgotten", id),
	}, nil
}

func (s *Server) toolListMemories(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	var tags []string
	if tagsRaw, ok := args["tags"].([]interface{}); ok {
		for _, t := range tagsRaw {
			if ts, ok := t.(string); ok {
				tags = append(tags, ts)
			}
		}
	}

	memories, err := s.store.List(ctx, limit, tags)
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, len(memories))
	for i, mem := range memories {
		results[i] = map[string]interface{}{
			"id":         mem.ID,
			"content":    truncate(mem.Content, 200),
			"tags":       mem.Tags,
			"created_at": mem.CreatedAt.Format(time.RFC3339),
			"source":     "local",
		}
	}

	return map[string]interface{}{
		"count":    len(results),
		"memories": results,
	}, nil
}

func (s *Server) toolMemoryStats(ctx context.Context) (interface{}, error) {
	stats := s.GetMemoryStats()
	return stats, nil
}

func (s *Server) toolSessionContext(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	hint := ""
	if h, ok := args["hint"].(string); ok {
		hint = h
	}

	var sb strings.Builder
	seen := make(map[string]bool) // Deduplication

	sb.WriteString("# Session Context Loaded\n\n")

	// SECTION 1: Hint-based semantic search with recency boost
	// This is the primary context loading - uses blended scoring
	if hint != "" {
		sb.WriteString(fmt.Sprintf("## Relevant to: %s\n\n", hint))

		// Use blended recall: 60% semantic, 30% recency, 10% importance
		relevant, err := s.store.RecallWithRecencyBoost(ctx, hint, 10, memory.RecallOptions{
			SemanticWeight:       0.6,
			RecencyWeight:        0.3,
			ImportanceWeight:     0.1,
			RecencyHalfLifeHours: 72, // 3-day half-life for session context (more aggressive recency)
		})
		if err == nil && len(relevant) > 0 {
			count := 0
			for _, mem := range relevant {
				if mem.Similarity < 0.15 { // Slightly higher threshold since blended
					continue
				}
				if seen[mem.ID] {
					continue
				}
				seen[mem.ID] = true

				sb.WriteString(fmt.Sprintf("**[%.0f%% match]** ", mem.Similarity*100))
				content := mem.Content
				if len(content) > 400 {
					content = content[:400] + "..."
				}
				sb.WriteString(content)
				sb.WriteString("\n\n")

				count++
				if count >= 5 {
					break
				}
			}
		}
	}

	// SECTION 2: Recent Activity (guaranteed last 10 memories)
	// These ALWAYS appear regardless of hint match
	recent, _ := s.store.List(ctx, 10, nil)
	if len(recent) > 0 {
		sb.WriteString("## Recent Activity\n\n")
		count := 0
		for _, mem := range recent {
			if seen[mem.ID] {
				continue // Already shown in relevant section
			}
			seen[mem.ID] = true

			sb.WriteString(fmt.Sprintf("**%s**", mem.CreatedAt.Format("Jan 2 15:04")))
			if len(mem.Tags) > 0 {
				interestingTags := filterInterestingTags(mem.Tags)
				if len(interestingTags) > 0 {
					sb.WriteString(fmt.Sprintf(" [%s]", strings.Join(interestingTags, ", ")))
				}
			}
			sb.WriteString("\n")
			content := mem.Content
			if len(content) > 300 {
				content = content[:300] + "..."
			}
			sb.WriteString(content)
			sb.WriteString("\n\n")

			count++
			if count >= 8 {
				break
			}
		}
	}

	// SECTION 3: Important memories from last 7 days (guaranteed slots)
	// These surface critical/milestone/decision items even if not in recent or hint-matched
	important, _ := s.store.GetRecentImportant(ctx, 7*24*time.Hour, 10)
	unseenImportant := []*memory.Memory{}
	for _, mem := range important {
		if !seen[mem.ID] {
			unseenImportant = append(unseenImportant, mem)
			seen[mem.ID] = true
		}
	}
	if len(unseenImportant) > 0 {
		sb.WriteString("## Critical (Last 7 Days)\n\n")
		for _, mem := range unseenImportant {
			content := mem.Content
			if len(content) > 250 {
				content = content[:250] + "..."
			}
			sb.WriteString(fmt.Sprintf("- %s\n", content))
		}
		sb.WriteString("\n")
	}

	// SECTION 4: Key categories (only if not already covered)
	for _, tag := range []string{"decision", "milestone"} {
		tagged, _ := s.store.List(ctx, 5, []string{tag})
		unseen := []*memory.Memory{}
		for _, mem := range tagged {
			if !seen[mem.ID] {
				unseen = append(unseen, mem)
				seen[mem.ID] = true
			}
		}
		if len(unseen) > 0 {
			sb.WriteString(fmt.Sprintf("## %s\n\n", strings.Title(tag)))
			for _, mem := range unseen {
				if len(unseen) > 3 {
					unseen = unseen[:3]
				}
				content := mem.Content
				if len(content) > 250 {
					content = content[:250] + "..."
				}
				sb.WriteString(fmt.Sprintf("- %s\n", content))
			}
			sb.WriteString("\n")
		}
	}

	stats := s.GetMemoryStats()
	sb.WriteString(fmt.Sprintf("---\n*Loaded from %d memories | Last activity: %s*\n",
		stats.TotalMemories, stats.LastActivity))

	return map[string]interface{}{
		"context": sb.String(),
		"stats": map[string]interface{}{
			"total_memories": stats.TotalMemories,
			"last_activity":  stats.LastActivity,
		},
	}, nil
}

// filterInterestingTags removes noise tags from display
func filterInterestingTags(tags []string) []string {
	boring := map[string]bool{
		"conversation":  true,
		"auto-ingested": true,
		"assistant":     true,
		"user":          true,
	}
	result := []string{}
	for _, t := range tags {
		if !boring[t] {
			result = append(result, t)
		}
	}
	return result
}

// handleResourcesList returns available resources
func (s *Server) handleResourcesList(req *JSONRPCRequest) {
	resources := []map[string]interface{}{
		{
			"uri":         "phloem://memories/recent",
			"name":        "Recent Memories",
			"description": "List of most recent memories",
			"mimeType":    "application/json",
		},
		{
			"uri":         "phloem://memories/stats",
			"name":        "Memory Statistics",
			"description": "Statistics about the memory store",
			"mimeType":    "application/json",
		},
		{
			"uri":         "phloem://context/session",
			"name":        "Session Context",
			"description": "Pre-loaded context for session start - includes recent memories, key decisions, and relevant background",
			"mimeType":    "text/markdown",
		},
	}

	s.sendResult(req.ID, map[string]interface{}{"resources": resources})
}

// handleResourceRead reads a resource
func (s *Server) handleResourceRead(ctx context.Context, req *JSONRPCRequest) {
	var params struct {
		URI string `json:"uri"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "Invalid params", err.Error())
		return
	}

	var content interface{}
	var err error

	switch params.URI {
	case "phloem://memories/recent":
		content, err = s.toolListMemories(ctx, map[string]interface{}{"limit": float64(10)})
	case "phloem://memories/stats":
		content, err = s.toolMemoryStats(ctx)
	case "phloem://context/session":
		// Return session context as markdown
		contextMd, err := s.buildSessionContext(ctx)
		if err != nil {
			s.sendError(req.ID, -32603, "Internal error", err.Error())
			return
		}
		s.sendResult(req.ID, map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"uri":      params.URI,
					"mimeType": "text/markdown",
					"text":     contextMd,
				},
			},
		})
		return
	default:
		s.sendError(req.ID, -32602, "Unknown resource", params.URI)
		return
	}

	if err != nil {
		s.sendError(req.ID, -32603, "Internal error", err.Error())
		return
	}

	text, _ := json.MarshalIndent(content, "", "  ")
	s.sendResult(req.ID, map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"uri":      params.URI,
				"mimeType": "application/json",
				"text":     string(text),
			},
		},
	})
}

// handlePromptsList returns available prompts
func (s *Server) handlePromptsList(req *JSONRPCRequest) {
	prompts := []map[string]interface{}{
		{
			"name":        "with_memory",
			"description": "Enhance your prompt with relevant memories",
			"arguments": []map[string]interface{}{
				{
					"name":        "query",
					"description": "Your current task or question",
					"required":    true,
				},
			},
		},
	}

	s.sendResult(req.ID, map[string]interface{}{"prompts": prompts})
}

// handlePromptsGet returns a prompt with relevant memories injected
func (s *Server) handlePromptsGet(ctx context.Context, req *JSONRPCRequest) {
	var params struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "Invalid params", err.Error())
		return
	}

	if params.Name != "with_memory" {
		s.sendError(req.ID, -32602, "Unknown prompt", params.Name)
		return
	}

	query := params.Arguments["query"]
	if query == "" {
		s.sendError(req.ID, -32602, "Missing required argument", "query")
		return
	}

	// Recall relevant memories
	var memoryContext string
	memories, err := s.store.Recall(ctx, query, 5, nil)
	if err == nil && len(memories) > 0 {
		memoryContext = "Relevant memories:\n"
		for _, mem := range memories {
			memoryContext += fmt.Sprintf("- %s\n", mem.Content)
		}
		memoryContext += "\n"
	}

	messages := []map[string]interface{}{
		{
			"role": "user",
			"content": map[string]interface{}{
				"type": "text",
				"text": memoryContext + query,
			},
		},
	}

	s.sendResult(req.ID, map[string]interface{}{
		"description": "Query enhanced with relevant memories",
		"messages":    messages,
	})
}

// JSON-RPC types and helpers

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

func (s *Server) sendResult(id interface{}, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	data, _ := json.Marshal(resp)
	fmt.Println(string(data))
}

func (s *Server) sendError(id interface{}, code int, message, data string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	respData, _ := json.Marshal(resp)
	fmt.Println(string(respData))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// buildSessionContext creates a markdown summary for session preload
func (s *Server) buildSessionContext(ctx context.Context) (string, error) {
	var sb strings.Builder

	sb.WriteString("# Phloem Session Context\n\n")
	sb.WriteString("*Auto-loaded memory context for this session*\n\n")

	// Get recent memories (last 24 hours of activity)
	recent, err := s.store.List(ctx, 10, nil)
	if err != nil {
		return "", err
	}

	if len(recent) > 0 {
		sb.WriteString("## Recent Context\n\n")
		for _, mem := range recent {
			// Format each memory
			sb.WriteString(fmt.Sprintf("**%s**", mem.CreatedAt.Format("Jan 2 15:04")))
			if len(mem.Tags) > 0 {
				sb.WriteString(fmt.Sprintf(" [%s]", strings.Join(mem.Tags, ", ")))
			}
			sb.WriteString("\n")

			// Truncate long content
			content := mem.Content
			if len(content) > 500 {
				content = content[:500] + "..."
			}
			sb.WriteString(content)
			sb.WriteString("\n\n")
		}
	}

	// Get key decisions
	decisions, _ := s.store.List(ctx, 5, []string{"decision"})
	if len(decisions) > 0 {
		sb.WriteString("## Key Decisions\n\n")
		for _, mem := range decisions {
			content := mem.Content
			if len(content) > 300 {
				content = content[:300] + "..."
			}
			sb.WriteString(fmt.Sprintf("- %s\n", content))
		}
		sb.WriteString("\n")
	}

	// Get critical/priority items
	critical, _ := s.store.List(ctx, 5, []string{"critical"})
	if len(critical) > 0 {
		sb.WriteString("## Critical Items\n\n")
		for _, mem := range critical {
			content := mem.Content
			if len(content) > 300 {
				content = content[:300] + "..."
			}
			sb.WriteString(fmt.Sprintf("- %s\n", content))
		}
		sb.WriteString("\n")
	}

	// Get architecture notes
	arch, _ := s.store.List(ctx, 3, []string{"architecture"})
	if len(arch) > 0 {
		sb.WriteString("## Architecture Notes\n\n")
		for _, mem := range arch {
			content := mem.Content
			if len(content) > 400 {
				content = content[:400] + "..."
			}
			sb.WriteString(content)
			sb.WriteString("\n\n")
		}
	}

	// SECTION 2: Developer Identity (Section 32)
	// Inject developer profile, coding styles, and preferences
	identity, _ := s.store.GetIdentityProfile(ctx)
	if len(identity) > 0 {
		sb.WriteString("## Developer Identity & Style\n\n")
		for _, mem := range identity {
			sb.WriteString(fmt.Sprintf("- %s\n", mem.Content))
		}
		sb.WriteString("\n")
	}

	// Stats
	stats := s.GetMemoryStats()
	sb.WriteString(fmt.Sprintf("---\n*%d total memories | Last activity: %s*\n",
		stats.TotalMemories, stats.LastActivity))

	return sb.String(), nil
}

// ============================================================================
// Citation Tools
// ============================================================================

func (s *Server) toolAddCitation(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	memoryID, ok := args["memory_id"].(string)
	if !ok || memoryID == "" {
		return nil, fmt.Errorf("memory_id is required")
	}

	filePath, ok := args["file_path"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("file_path is required")
	}

	// Optional parameters
	startLine := 0
	if sl, ok := args["start_line"].(float64); ok {
		startLine = int(sl)
	}

	endLine := 0
	if el, ok := args["end_line"].(float64); ok {
		endLine = int(el)
	}

	content := ""
	if c, ok := args["content"].(string); ok {
		content = c
	}

	// Get current git commit SHA if available
	commitSHA := ""
	// Try to get git SHA from the file's directory
	if filePath != "" {
		// This is a simple implementation - in production you'd use git commands
		commitSHA = "" // Could be enhanced to actually get git SHA
	}

	citation, err := s.store.AddCitation(ctx, memoryID, filePath, startLine, endLine, commitSHA, content)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status":   "citation_added",
		"citation": citation,
		"message":  fmt.Sprintf("Citation added linking memory %s to %s", memoryID, filePath),
	}, nil
}

func (s *Server) toolVerifyCitation(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	citationID, ok := args["citation_id"].(string)
	if !ok || citationID == "" {
		return nil, fmt.Errorf("citation_id is required")
	}

	citation, valid, err := s.store.VerifyCitation(ctx, citationID)
	if err != nil {
		return nil, err
	}

	status := "invalid"
	if valid {
		status = "valid"
	}

	return map[string]interface{}{
		"status":     status,
		"citation":   citation,
		"confidence": citation.Confidence,
		"message": fmt.Sprintf("Citation %s is %s (confidence: %.0f%%)",
			citationID, status, citation.Confidence*100),
	}, nil
}

func (s *Server) toolGetCitations(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	memoryID, ok := args["memory_id"].(string)
	if !ok || memoryID == "" {
		return nil, fmt.Errorf("memory_id is required")
	}

	citations, err := s.store.GetCitations(ctx, memoryID)
	if err != nil {
		return nil, err
	}

	// Calculate aggregate confidence
	confidence, _ := s.store.GetMemoryConfidence(ctx, memoryID)

	return map[string]interface{}{
		"memory_id":            memoryID,
		"citations":            citations,
		"count":                len(citations),
		"aggregate_confidence": confidence,
	}, nil
}

// toolVerifyMemory verifies all citations for a memory and updates confidence scores
func (s *Server) toolVerifyMemory(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	memoryID, ok := args["memory_id"].(string)
	if !ok || memoryID == "" {
		return nil, fmt.Errorf("memory_id is required")
	}

	// Get all citations for this memory
	citations, err := s.store.GetCitations(ctx, memoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get citations: %w", err)
	}

	if len(citations) == 0 {
		return map[string]interface{}{
			"memory_id":  memoryID,
			"status":     "no_citations",
			"message":    "Memory has no citations to verify",
			"confidence": 1.0, // No citations = full confidence
		}, nil
	}

	// Verify each citation
	verified := 0
	invalid := 0
	totalConfidence := 0.0

	for _, citation := range citations {
		_, valid, err := s.store.VerifyCitation(ctx, citation.ID)
		if err != nil {
			continue
		}

		if valid {
			verified++
		} else {
			invalid++
		}

		// Re-fetch citation to get updated confidence
		updatedCitations, _ := s.store.GetCitations(ctx, memoryID)
		for _, c := range updatedCitations {
			if c.ID == citation.ID {
				totalConfidence += c.Confidence
				break
			}
		}
	}

	// Calculate aggregate confidence
	confidence, _ := s.store.GetMemoryConfidence(ctx, memoryID)

	return map[string]interface{}{
		"memory_id":  memoryID,
		"status":     "verified",
		"verified":   verified,
		"invalid":    invalid,
		"total":      len(citations),
		"confidence": confidence,
		"message":    fmt.Sprintf("Verified %d citation(s): %d valid, %d invalid (confidence: %.0f%%)", len(citations), verified, invalid, confidence*100),
	}, nil
}

func (s *Server) toolCausalQuery(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	memoryID, ok := args["memory_id"].(string)
	if !ok || memoryID == "" {
		return nil, fmt.Errorf("memory_id is required")
	}
	queryType, _ := args["query_type"].(string)
	if queryType == "" {
		queryType = "neighbors"
	}

	switch strings.ToLower(queryType) {
	case "neighbors":
		memories, err := s.store.CausalNeighbors(ctx, memoryID)
		if err != nil {
			return nil, fmt.Errorf("causal neighbors: %w", err)
		}
		items := make([]map[string]interface{}, len(memories))
		for i, m := range memories {
			items[i] = map[string]interface{}{
				"id":         m.ID,
				"content":    m.Content,
				"tags":       m.Tags,
				"created_at": m.CreatedAt.Format(time.RFC3339),
			}
		}
		return map[string]interface{}{
			"memory_id":  memoryID,
			"query_type": "neighbors",
			"count":      len(items),
			"memories":   items,
		}, nil
	case "affected":
		ids, err := s.store.AffectedIfChanged(ctx, memoryID)
		if err != nil {
			return nil, fmt.Errorf("affected if changed: %w", err)
		}
		return map[string]interface{}{
			"memory_id":  memoryID,
			"query_type": "affected",
			"count":      len(ids),
			"memory_ids": ids,
		}, nil
	default:
		return nil, fmt.Errorf("query_type must be 'neighbors' or 'affected', got %q", queryType)
	}
}

// toolCompose: accepts either query_a+query_b or queries ([]string). Uses stage2 Store.Compose.
func (s *Server) toolCompose(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	var queries []string
	if raw, ok := args["queries"].([]interface{}); ok && len(raw) > 0 {
		for _, q := range raw {
			if qs, ok := q.(string); ok && qs != "" {
				queries = append(queries, qs)
			}
		}
	}
	if len(queries) == 0 {
		queryA, okA := args["query_a"].(string)
		queryB, okB := args["query_b"].(string)
		if okA && okB && queryA != "" && queryB != "" {
			queries = []string{queryA, queryB}
		}
	}
	if len(queries) == 0 {
		return nil, fmt.Errorf("queries (array of strings) or query_a and query_b are required")
	}
	limit := 10
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}
	composed, err := s.store.Compose(ctx, queries, limit)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]interface{}, len(composed.Memories))
	for i, m := range composed.Memories {
		items[i] = map[string]interface{}{
			"id":         m.ID,
			"content":    m.Content,
			"tags":       m.Tags,
			"context":    m.Context,
			"similarity": m.Similarity,
			"created_at": m.CreatedAt.Format(time.RFC3339),
		}
	}
	return map[string]interface{}{
		"explanation": composed.Explanation,
		"count":       len(items),
		"memories":    items,
	}, nil
}

func (s *Server) toolPrefetch(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	hint := ""
	if h, ok := args["context_hint"].(string); ok {
		hint = h
	}
	limit := 5
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}
	memories, err := s.store.PrefetchSuggest(ctx, hint, limit)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]interface{}, len(memories))
	for i, m := range memories {
		content := m.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		items[i] = map[string]interface{}{
			"id":         m.ID,
			"content":    content,
			"tags":       m.Tags,
			"created_at": m.CreatedAt.Format(time.RFC3339),
		}
	}
	return map[string]interface{}{
		"context_hint": hint,
		"count":        len(items),
		"suggestions":  items,
	}, nil
}

// toolPrefetchSuggest (Stage 2): suggest memories to preload given current context.
func (s *Server) toolPrefetchSuggest(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	ctxStr, ok := args["context"].(string)
	if !ok || ctxStr == "" {
		return nil, fmt.Errorf("context is required")
	}
	limit := 5
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}
	if limit > 20 {
		limit = 20
	}
	memories, err := s.store.PrefetchSuggest(ctx, ctxStr, limit)
	if err != nil {
		return nil, err
	}
	results := make([]map[string]interface{}, len(memories))
	for i, mem := range memories {
		results[i] = map[string]interface{}{
			"id":         mem.ID,
			"content":    mem.Content,
			"tags":       mem.Tags,
			"similarity": mem.Similarity,
		}
	}
	return map[string]interface{}{
		"context":     ctxStr,
		"count":       len(results),
		"suggestions": results,
	}, nil
}
