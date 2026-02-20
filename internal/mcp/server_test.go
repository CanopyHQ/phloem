package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// captureOutput redirects stdout during test and returns captured content
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// setupTestServer creates a server with a temp data directory
func setupTestServer(t *testing.T) (*Server, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "phloem-mcp-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	originalDataDir := os.Getenv("PHLOEM_DATA_DIR")
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)

	// Suppress stderr output during tests
	oldStderr := os.Stderr
	os.Stderr, _ = os.Open(os.DevNull)

	server, err := NewServer()

	os.Stderr = oldStderr

	if err != nil {
		os.RemoveAll(tmpDir)
		os.Setenv("PHLOEM_DATA_DIR", originalDataDir)
		t.Fatalf("failed to create server: %v", err)
	}

	cleanup := func() {
		server.Stop()
		os.RemoveAll(tmpDir)
		os.Setenv("PHLOEM_DATA_DIR", originalDataDir)
	}

	return server, cleanup
}

// =============================================================================
// Server Creation Tests
// =============================================================================

func TestNewServer_Basic(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "phloem-mcp-basic-*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	origDir := os.Getenv("PHLOEM_DATA_DIR")
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer func() { os.Setenv("PHLOEM_DATA_DIR", origDir) }()

	oldStderr := os.Stderr
	os.Stderr, _ = os.Open(os.DevNull)
	server, err := NewServer()
	os.Stderr = oldStderr

	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer server.Stop()

	if server.store == nil {
		t.Error("expected non-nil store")
	}
}

func TestNewServer(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	if server == nil {
		t.Fatal("expected non-nil server")
	}
	if server.store == nil {
		t.Error("expected non-nil store")
	}
}

// =============================================================================
// Initialize Tests
// =============================================================================

func TestHandleInitialize(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}

	output := captureOutput(func() {
		server.handleRequest(req)
	})

	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("result is not a map")
	}

	// Check protocol version
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("unexpected protocol version: %v", result["protocolVersion"])
	}

	// Check capabilities
	caps, ok := result["capabilities"].(map[string]interface{})
	if !ok {
		t.Error("capabilities missing")
	}
	if caps["tools"] == nil {
		t.Error("tools capability missing")
	}

	// Check server info
	info, ok := result["serverInfo"].(map[string]interface{})
	if !ok {
		t.Error("serverInfo missing")
	}
	if info["name"] != "phloem-mcp" {
		t.Errorf("unexpected server name: %v", info["name"])
	}
}

// =============================================================================
// Tools List Tests
// =============================================================================

func TestHandleToolsList(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}

	output := captureOutput(func() {
		server.handleRequest(req)
	})

	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("result is not a map")
	}

	tools, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatal("tools is not an array")
	}

	// Should include the core tools and any extended tools
	expectedTools := map[string]bool{
		"remember":                  false,
		"recall":                    false,
		"forget":                    false,
		"list_memories":             false,
		"memory_stats":              false,
		"session_context":           false,
		"add_citation":    false,
		"verify_citation": false,
		"get_citations":   false,
		"verify_memory":   false,
	}

	for _, tool := range tools {
		toolMap := tool.(map[string]interface{})
		name := toolMap["name"].(string)
		expectedTools[name] = true
	}

	for name, found := range expectedTools {
		if !found {
			t.Errorf("tool '%s' not found in tools list", name)
		}
	}
}

func TestToolsHaveValidSchema(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}

	output := captureOutput(func() {
		server.handleRequest(req)
	})

	var resp JSONRPCResponse
	json.Unmarshal([]byte(output), &resp)

	result := resp.Result.(map[string]interface{})
	tools := result["tools"].([]interface{})

	for _, tool := range tools {
		toolMap := tool.(map[string]interface{})
		name := toolMap["name"].(string)

		// Check required fields
		if toolMap["description"] == nil {
			t.Errorf("tool '%s' missing description", name)
		}
		if toolMap["inputSchema"] == nil {
			t.Errorf("tool '%s' missing inputSchema", name)
		}

		// Validate inputSchema structure
		schema := toolMap["inputSchema"].(map[string]interface{})
		if schema["type"] != "object" {
			t.Errorf("tool '%s' schema type should be 'object'", name)
		}
	}
}

// =============================================================================
// Tool Call Tests - Remember
// =============================================================================

func TestToolCall_Remember(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	params := map[string]interface{}{
		"name": "remember",
		"arguments": map[string]interface{}{
			"content": "Test memory content",
			"tags":    []interface{}{"test", "example"},
			"context": "test context",
		},
	}
	paramsJSON, _ := json.Marshal(params)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  paramsJSON,
	}

	output := captureOutput(func() {
		server.handleRequest(req)
	})

	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}

	// Check result contains content with status
	result := resp.Result.(map[string]interface{})
	content := result["content"].([]interface{})
	if len(content) == 0 {
		t.Error("expected content in result")
	}

	textContent := content[0].(map[string]interface{})
	text := textContent["text"].(string)
	if !strings.Contains(text, "remembered") {
		t.Errorf("expected 'remembered' in response, got: %s", text)
	}
}

func TestToolCall_Remember_MissingContent(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	params := map[string]interface{}{
		"name": "remember",
		"arguments": map[string]interface{}{
			"tags": []interface{}{"test"},
		},
	}
	paramsJSON, _ := json.Marshal(params)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  paramsJSON,
	}

	output := captureOutput(func() {
		server.handleRequest(req)
	})

	var resp JSONRPCResponse
	json.Unmarshal([]byte(output), &resp)

	result := resp.Result.(map[string]interface{})
	if result["isError"] != true {
		t.Error("expected isError for missing content")
	}
}

// =============================================================================
// Tool Call Tests - Recall
// =============================================================================

func TestToolCall_Recall(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// First, add a memory
	ctx := context.Background()
	server.store.Remember(ctx, "Go is a great programming language", []string{"code"}, "")

	params := map[string]interface{}{
		"name": "recall",
		"arguments": map[string]interface{}{
			"query": "programming",
			"limit": 5,
		},
	}
	paramsJSON, _ := json.Marshal(params)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  paramsJSON,
	}

	output := captureOutput(func() {
		server.handleRequest(req)
	})

	var resp JSONRPCResponse
	json.Unmarshal([]byte(output), &resp)

	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}

	result := resp.Result.(map[string]interface{})
	content := result["content"].([]interface{})
	textContent := content[0].(map[string]interface{})
	text := textContent["text"].(string)

	// Should contain the recalled memory
	if !strings.Contains(text, "Go is a great programming language") {
		t.Errorf("expected memory content in recall result: %s", text)
	}
}

func TestToolRecall_LocalStore(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	server.store.Remember(ctx, "local fallback memory", []string{"tag1"}, "")

	// Recall with tags hits store.Recall(ctx, query, limit, tags)
	paramsWithTags := map[string]interface{}{
		"name": "recall",
		"arguments": map[string]interface{}{
			"query": "fallback",
			"limit": 5.0,
			"tags":  []interface{}{"tag1"},
		},
	}
	paramsJSON, _ := json.Marshal(paramsWithTags)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("recall with tags: %v", resp.Error)
	}

	// Recall without tags hits RecallWithRecencyBoost path
	paramsNoTags := map[string]interface{}{
		"name": "recall",
		"arguments": map[string]interface{}{
			"query": "local",
			"limit": 3.0,
		},
	}
	paramsJSON, _ = json.Marshal(paramsNoTags)
	req = &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output = captureOutput(func() { server.handleRequest(req) })
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("recall no tags: %v", resp.Error)
	}
}

func TestToolCall_Recall_MissingQuery(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	params := map[string]interface{}{
		"name": "recall",
		"arguments": map[string]interface{}{
			"limit": 5,
		},
	}
	paramsJSON, _ := json.Marshal(params)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  paramsJSON,
	}

	output := captureOutput(func() {
		server.handleRequest(req)
	})

	var resp JSONRPCResponse
	json.Unmarshal([]byte(output), &resp)

	result := resp.Result.(map[string]interface{})
	if result["isError"] != true {
		t.Error("expected isError for missing query")
	}
}

func TestToolCall_Recall_WithTagFilter(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	server.store.Remember(ctx, "Memory with code tag", []string{"code"}, "")
	server.store.Remember(ctx, "Memory with design tag", []string{"design"}, "")

	params := map[string]interface{}{
		"name": "recall",
		"arguments": map[string]interface{}{
			"query": "memory",
			"tags":  []interface{}{"code"},
		},
	}
	paramsJSON, _ := json.Marshal(params)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  paramsJSON,
	}

	output := captureOutput(func() {
		server.handleRequest(req)
	})

	var resp JSONRPCResponse
	json.Unmarshal([]byte(output), &resp)

	result := resp.Result.(map[string]interface{})
	content := result["content"].([]interface{})
	textContent := content[0].(map[string]interface{})
	text := textContent["text"].(string)

	// Should only contain code-tagged memory
	if !strings.Contains(text, "code tag") {
		t.Errorf("expected code-tagged memory in result: %s", text)
	}
	if strings.Contains(text, "design tag") {
		t.Errorf("should not contain design-tagged memory: %s", text)
	}
}

// =============================================================================
// Tool Call Tests - Forget
// =============================================================================

func TestToolCall_Forget(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// First, add a memory
	ctx := context.Background()
	mem, _ := server.store.Remember(ctx, "Memory to forget", nil, "")

	params := map[string]interface{}{
		"name": "forget",
		"arguments": map[string]interface{}{
			"id": mem.ID,
		},
	}
	paramsJSON, _ := json.Marshal(params)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  paramsJSON,
	}

	output := captureOutput(func() {
		server.handleRequest(req)
	})

	var resp JSONRPCResponse
	json.Unmarshal([]byte(output), &resp)

	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}

	result := resp.Result.(map[string]interface{})
	content := result["content"].([]interface{})
	textContent := content[0].(map[string]interface{})
	text := textContent["text"].(string)

	if !strings.Contains(text, "forgotten") {
		t.Errorf("expected 'forgotten' in response: %s", text)
	}

	// Verify memory is gone
	count, _ := server.store.Count(ctx)
	if count != 0 {
		t.Error("memory should have been deleted")
	}
}

func TestToolCall_Forget_MissingID(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	params := map[string]interface{}{
		"name":      "forget",
		"arguments": map[string]interface{}{},
	}
	paramsJSON, _ := json.Marshal(params)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  paramsJSON,
	}

	output := captureOutput(func() {
		server.handleRequest(req)
	})

	var resp JSONRPCResponse
	json.Unmarshal([]byte(output), &resp)

	result := resp.Result.(map[string]interface{})
	if result["isError"] != true {
		t.Error("expected isError for missing id")
	}
}

// =============================================================================
// Tool Call Tests - List Memories
// =============================================================================

func TestToolCall_ListMemories(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	server.store.Remember(ctx, "Memory 1", []string{"tag1"}, "")
	server.store.Remember(ctx, "Memory 2", []string{"tag2"}, "")

	params := map[string]interface{}{
		"name": "list_memories",
		"arguments": map[string]interface{}{
			"limit": 10,
		},
	}
	paramsJSON, _ := json.Marshal(params)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  paramsJSON,
	}

	output := captureOutput(func() {
		server.handleRequest(req)
	})

	var resp JSONRPCResponse
	json.Unmarshal([]byte(output), &resp)

	result := resp.Result.(map[string]interface{})
	content := result["content"].([]interface{})
	textContent := content[0].(map[string]interface{})
	text := textContent["text"].(string)

	if !strings.Contains(text, `"count": 2`) {
		t.Errorf("expected count: 2 in response: %s", text)
	}
}

func TestToolListMemories_LocalStore(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	server.store.Remember(ctx, "local list memory", []string{"milestone"}, "")

	params := map[string]interface{}{
		"name":      "list_memories",
		"arguments": map[string]interface{}{"limit": 5.0, "tags": []interface{}{"milestone"}},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("list_memories local fallback: %v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	content := result["content"].([]interface{})
	if len(content) == 0 {
		t.Fatal("expected content")
	}
	text := content[0].(map[string]interface{})["text"].(string)
	if !strings.Contains(text, `"count": 1`) && !strings.Contains(text, `"count":1`) {
		t.Errorf("expected count 1 in response: %s", text[:min(200, len(text))])
	}
}

// =============================================================================
// Tool Call Tests - Memory Stats
// =============================================================================

func TestToolCall_MemoryStats(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	server.store.Remember(ctx, "Test memory", nil, "")

	params := map[string]interface{}{
		"name":      "memory_stats",
		"arguments": map[string]interface{}{},
	}
	paramsJSON, _ := json.Marshal(params)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  paramsJSON,
	}

	output := captureOutput(func() {
		server.handleRequest(req)
	})

	var resp JSONRPCResponse
	json.Unmarshal([]byte(output), &resp)

	result := resp.Result.(map[string]interface{})
	content := result["content"].([]interface{})
	textContent := content[0].(map[string]interface{})
	text := textContent["text"].(string)

	// Should contain stats fields
	if !strings.Contains(text, "total_memories") {
		t.Errorf("expected total_memories in stats: %s", text)
	}
	if !strings.Contains(text, "database_size") {
		t.Errorf("expected database_size in stats: %s", text)
	}
}

// =============================================================================
// Tool Call Tests - Unknown Tool
// =============================================================================

func TestToolCall_UnknownTool(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	params := map[string]interface{}{
		"name":      "unknown_tool",
		"arguments": map[string]interface{}{},
	}
	paramsJSON, _ := json.Marshal(params)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  paramsJSON,
	}

	output := captureOutput(func() {
		server.handleRequest(req)
	})

	var resp JSONRPCResponse
	json.Unmarshal([]byte(output), &resp)

	if resp.Error == nil {
		t.Error("expected error for unknown tool")
	}
	if resp.Error.Code != -32602 {
		t.Errorf("expected error code -32602, got %d", resp.Error.Code)
	}
}

// =============================================================================
// Resources Tests
// =============================================================================

func TestHandleResourcesList(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "resources/list",
	}

	output := captureOutput(func() {
		server.handleRequest(req)
	})

	var resp JSONRPCResponse
	json.Unmarshal([]byte(output), &resp)

	result := resp.Result.(map[string]interface{})
	resources := result["resources"].([]interface{})

	// Should have recent and stats resources
	expectedURIs := map[string]bool{
		"phloem://memories/recent": false,
		"phloem://memories/stats":  false,
	}

	for _, res := range resources {
		resMap := res.(map[string]interface{})
		uri := resMap["uri"].(string)
		expectedURIs[uri] = true
	}

	for uri, found := range expectedURIs {
		if !found {
			t.Errorf("resource '%s' not found", uri)
		}
	}
}

func TestHandleResourceRead_RecentMemories(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	server.store.Remember(ctx, "Test memory for resource", nil, "")

	params := map[string]interface{}{
		"uri": "phloem://memories/recent",
	}
	paramsJSON, _ := json.Marshal(params)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "resources/read",
		Params:  paramsJSON,
	}

	output := captureOutput(func() {
		server.handleRequest(req)
	})

	var resp JSONRPCResponse
	json.Unmarshal([]byte(output), &resp)

	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}

	result := resp.Result.(map[string]interface{})
	contents := result["contents"].([]interface{})
	if len(contents) == 0 {
		t.Error("expected contents in response")
	}
}

func TestHandleResourceRead_Stats(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	params := map[string]interface{}{
		"uri": "phloem://memories/stats",
	}
	paramsJSON, _ := json.Marshal(params)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "resources/read",
		Params:  paramsJSON,
	}

	output := captureOutput(func() {
		server.handleRequest(req)
	})

	var resp JSONRPCResponse
	json.Unmarshal([]byte(output), &resp)

	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
}

func TestHandleResourceRead_InvalidParams(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Invalid JSON so handleResourceRead returns Invalid params
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "resources/read", Params: json.RawMessage(`{`)}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error == nil {
		t.Error("expected error for invalid params")
	}
}

func TestHandleResourceRead_NoURI_UnknownResource(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	params := map[string]interface{}{} // empty => URI ""
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "resources/read", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error == nil {
		t.Error("expected error for unknown resource")
	}
}

func TestHandleResourceRead_UnknownURI(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	params := map[string]interface{}{
		"uri": "phloem://unknown/resource",
	}
	paramsJSON, _ := json.Marshal(params)

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "resources/read",
		Params:  paramsJSON,
	}

	output := captureOutput(func() {
		server.handleRequest(req)
	})

	var resp JSONRPCResponse
	json.Unmarshal([]byte(output), &resp)

	if resp.Error == nil {
		t.Error("expected error for unknown resource")
	}
}

func TestHandleResourceRead_SessionContext(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	server.store.Remember(ctx, "Session context memory", nil, "")

	params := map[string]interface{}{
		"uri": "phloem://context/session",
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "resources/read", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	contents := result["contents"].([]interface{})
	if len(contents) == 0 {
		t.Fatal("expected contents")
	}
	text := contents[0].(map[string]interface{})["text"].(string)
	if !strings.Contains(text, "Session Context") && !strings.Contains(text, "Phloem") {
		t.Errorf("expected session context body: %s", text[:min(150, len(text))])
	}
}

func TestToolCall_SessionContext_WithBoringTags(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	server.store.Remember(ctx, "Memory with conversation tag", []string{"conversation", "code"}, "")

	params := map[string]interface{}{
		"name":      "session_context",
		"arguments": map[string]interface{}{"hint": "code"},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	content := result["content"].([]interface{})
	text := content[0].(map[string]interface{})["text"].(string)
	if !strings.Contains(text, "Session Context") && !strings.Contains(text, "Relevant") && !strings.Contains(text, "Recent") {
		t.Errorf("expected session context: %s", text[:min(200, len(text))])
	}
}

func TestToolCall_VerifyMemory_WithCitations(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	mem, _ := server.store.Remember(ctx, "Memory with citation", nil, "")
	tmpDir := t.TempDir()
	fpath := tmpDir + "/f.go"
	os.WriteFile(fpath, []byte("line1\nline2\nline3\n"), 0644)
	cite, _ := server.store.AddCitation(ctx, mem.ID, fpath, 1, 2, "", "line1\nline2")

	params := map[string]interface{}{
		"name":      "verify_memory",
		"arguments": map[string]interface{}{"memory_id": mem.ID},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	content := result["content"].([]interface{})
	text := content[0].(map[string]interface{})["text"].(string)
	if !strings.Contains(text, "verified") && !strings.Contains(text, "invalid") {
		t.Errorf("expected verified/invalid in response: %s", text)
	}
	_ = cite
}

// =============================================================================
// Prompts Tests
// =============================================================================

func TestHandlePromptsList(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "prompts/list",
	}

	output := captureOutput(func() {
		server.handleRequest(req)
	})

	var resp JSONRPCResponse
	json.Unmarshal([]byte(output), &resp)

	result := resp.Result.(map[string]interface{})
	prompts := result["prompts"].([]interface{})

	if len(prompts) == 0 {
		t.Error("expected at least one prompt")
	}

	// Check with_memory prompt exists
	found := false
	for _, p := range prompts {
		prompt := p.(map[string]interface{})
		if prompt["name"] == "with_memory" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'with_memory' prompt")
	}
}

// =============================================================================
// Error Handling Tests
// =============================================================================

func TestUnknownMethod(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "unknown/method",
	}

	output := captureOutput(func() {
		server.handleRequest(req)
	})

	var resp JSONRPCResponse
	json.Unmarshal([]byte(output), &resp)

	if resp.Error == nil {
		t.Error("expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("expected error code -32601 (Method not found), got %d", resp.Error.Code)
	}
}

func TestInvalidParams(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := &JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params:  json.RawMessage(`"invalid params"`), // Should be object
	}

	output := captureOutput(func() {
		server.handleRequest(req)
	})

	var resp JSONRPCResponse
	json.Unmarshal([]byte(output), &resp)

	if resp.Error == nil {
		t.Error("expected error for invalid params")
	}
}

// =============================================================================
// GetMemoryStats Tests
// =============================================================================

func TestGetMemoryStats(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	server.store.Remember(ctx, "Test memory", nil, "")

	stats := server.GetMemoryStats()

	if stats.TotalMemories != 1 {
		t.Errorf("expected 1 memory, got %d", stats.TotalMemories)
	}
	if stats.DatabaseSize == "" {
		t.Error("expected non-empty database size")
	}
	if stats.LastActivity == "never" {
		t.Error("expected last activity after adding memory")
	}
}

func TestGetMemoryStats_Empty(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	stats := server.GetMemoryStats()

	if stats.TotalMemories != 0 {
		t.Errorf("expected 0 memories, got %d", stats.TotalMemories)
	}
}

// =============================================================================
// Tool Call Tests - session_context, citations, compose, etc.
// =============================================================================

func TestToolCall_SessionContext(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	server.store.Remember(ctx, "session hint memory", []string{"a"}, "")

	params := map[string]interface{}{
		"name":      "session_context",
		"arguments": map[string]interface{}{"hint": "hint"},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	content := result["content"].([]interface{})
	if len(content) == 0 {
		t.Fatal("expected content")
	}
	text := content[0].(map[string]interface{})["text"].(string)
	if !strings.Contains(text, "Session Context") {
		t.Errorf("expected session context body: %s", text[:min(200, len(text))])
	}
}

func TestFilterInterestingTags(t *testing.T) {
	// Direct unit test for filterInterestingTags
	tests := []struct {
		tags   []string
		expect []string
	}{
		{nil, []string{}},
		{[]string{}, []string{}},
		{[]string{"conversation", "user"}, []string{}},
		{[]string{"decision", "conversation"}, []string{"decision"}},
		{[]string{"auto-ingested", "assistant", "task"}, []string{"task"}},
		{[]string{"a", "b"}, []string{"a", "b"}},
	}
	for _, tt := range tests {
		got := filterInterestingTags(tt.tags)
		if len(got) != len(tt.expect) {
			t.Errorf("filterInterestingTags(%v) = %v, want %v", tt.tags, got, tt.expect)
			continue
		}
		for i := range got {
			if got[i] != tt.expect[i] {
				t.Errorf("filterInterestingTags(%v) = %v, want %v", tt.tags, got, tt.expect)
				break
			}
		}
	}
}

func TestToolCall_SessionContext_FilterInterestingTags(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	// Memory with boring tag (conversation) and interesting tag (decision) to cover filterInterestingTags
	server.store.Remember(ctx, "Decision: use Go for backend", []string{"conversation", "decision"}, "")

	params := map[string]interface{}{
		"name":      "session_context",
		"arguments": map[string]interface{}{"hint": "decision"},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	content := result["content"].([]interface{})
	if len(content) == 0 {
		t.Fatal("expected content")
	}
	text := content[0].(map[string]interface{})["text"].(string)
	// filterInterestingTags keeps "decision", filters "conversation"; output should contain decision
	if !strings.Contains(text, "decision") {
		t.Errorf("expected decision in session context (filterInterestingTags): %s", text[:min(300, len(text))])
	}
}

func TestToolCall_AddCitation(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	mem, _ := server.store.Remember(ctx, "cite me", nil, "")

	params := map[string]interface{}{
		"name": "add_citation",
		"arguments": map[string]interface{}{
			"memory_id":  mem.ID,
			"file_path":  "/some/file.go",
			"start_line": 1.0, "end_line": 5.0,
			"content": "snippet",
		},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	content := result["content"].([]interface{})
	text := content[0].(map[string]interface{})["text"].(string)
	if !strings.Contains(text, "citation_added") {
		t.Errorf("expected citation_added: %s", text)
	}
}

func TestToolCall_VerifyCitation(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	mem, _ := server.store.Remember(ctx, "x", nil, "")
	cite, _ := server.store.AddCitation(ctx, mem.ID, "/f", 1, 2, "", "snip")

	params := map[string]interface{}{
		"name":      "verify_citation",
		"arguments": map[string]interface{}{"citation_id": cite.ID},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	content := result["content"].([]interface{})
	text := content[0].(map[string]interface{})["text"].(string)
	if !strings.Contains(text, "valid") && !strings.Contains(text, "invalid") {
		t.Errorf("expected valid/invalid: %s", text)
	}
}

func TestToolCall_GetCitations(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	mem, _ := server.store.Remember(ctx, "y", nil, "")

	params := map[string]interface{}{
		"name":      "get_citations",
		"arguments": map[string]interface{}{"memory_id": mem.ID},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	content := result["content"].([]interface{})
	text := content[0].(map[string]interface{})["text"].(string)
	if !strings.Contains(text, "memory_id") {
		t.Errorf("expected memory_id in response: %s", text)
	}
}

func TestToolCall_Recall_EmptyQuery_Error(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	params := map[string]interface{}{
		"name":      "recall",
		"arguments": map[string]interface{}{"query": ""},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	// Tool errors are returned as result with isError: true and content "Error: ..."
	result, _ := resp.Result.(map[string]interface{})
	if resp.Error != nil {
		return
	}
	if isErr, _ := result["isError"].(bool); isErr {
		return
	}
	if content, ok := result["content"].([]interface{}); ok && len(content) > 0 {
		if text, ok := content[0].(map[string]interface{})["text"].(string); ok && strings.Contains(text, "query is required") {
			return
		}
	}
	t.Error("expected error for empty query (resp.Error, isError, or content)")
}

func TestToolCall_Recall_WithTagsAndLimit(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	server.store.Remember(ctx, "Tagged recall target", []string{"decision"}, "")

	params := map[string]interface{}{
		"name":      "recall",
		"arguments": map[string]interface{}{"query": "recall target", "limit": 3.0, "tags": []interface{}{"decision"}},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	// Result is wrapped in content[0].text as JSON string
	content := result["content"].([]interface{})
	if len(content) == 0 {
		t.Fatal("expected content")
	}
	text := content[0].(map[string]interface{})["text"].(string)
	if !strings.Contains(text, "memories") && !strings.Contains(text, "query") {
		t.Errorf("expected memories or query in response: %s", text[:min(200, len(text))])
	}
}

func TestToolCall_ListMemories_WithLimitAndTags(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	server.store.Remember(ctx, "List with tags", []string{"milestone"}, "")

	params := map[string]interface{}{
		"name":      "list_memories",
		"arguments": map[string]interface{}{"limit": 5.0, "tags": []interface{}{"milestone"}},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	content := result["content"].([]interface{})
	if len(content) == 0 {
		t.Fatal("expected content")
	}
	text := content[0].(map[string]interface{})["text"].(string)
	if !strings.Contains(text, "count") && !strings.Contains(text, "memories") {
		t.Errorf("expected count/memories in response: %s", text[:min(150, len(text))])
	}
}

func TestToolCall_VerifyMemory(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	mem, _ := server.store.Remember(ctx, "z", nil, "")

	params := map[string]interface{}{
		"name":      "verify_memory",
		"arguments": map[string]interface{}{"memory_id": mem.ID},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	content := result["content"].([]interface{})
	text := content[0].(map[string]interface{})["text"].(string)
	if !strings.Contains(text, "no_citations") && !strings.Contains(text, "verified") {
		t.Errorf("expected no_citations or verified: %s", text)
	}
}

func TestToolCall_CausalQuery_Neighbors(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	mem, _ := server.store.Remember(ctx, "causal mem", nil, "")

	params := map[string]interface{}{
		"name":      "causal_query",
		"arguments": map[string]interface{}{"memory_id": mem.ID, "query_type": "neighbors"},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	content := result["content"].([]interface{})
	text := content[0].(map[string]interface{})["text"].(string)
	if !strings.Contains(text, "memory_id") {
		t.Errorf("expected memory_id: %s", text)
	}
}

func TestToolCall_CausalQuery_Affected(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	mem, _ := server.store.Remember(ctx, "affected mem", nil, "")

	params := map[string]interface{}{
		"name":      "causal_query",
		"arguments": map[string]interface{}{"memory_id": mem.ID, "query_type": "affected"},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
}

func TestToolCall_Compose(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	server.store.Remember(ctx, "compose a", nil, "")
	server.store.Remember(ctx, "compose b", nil, "")

	params := map[string]interface{}{
		"name": "compose",
		"arguments": map[string]interface{}{
			"query_a": "compose", "query_b": "compose",
			"limit": 5.0,
		},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
}

func TestToolCall_PrefetchSuggest(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	params := map[string]interface{}{
		"name":      "prefetch_suggest",
		"arguments": map[string]interface{}{"context": "test", "limit": 5.0},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
}

func TestToolCall_AddCitation_MissingMemoryID(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	params := map[string]interface{}{
		"name":      "add_citation",
		"arguments": map[string]interface{}{"file_path": "/f"},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	json.Unmarshal([]byte(output), &resp)
	result := resp.Result.(map[string]interface{})
	if result["isError"] != true {
		t.Error("expected isError for missing memory_id")
	}
}

func TestToolCall_Compose_EmptyQueries(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	params := map[string]interface{}{
		"name":      "compose",
		"arguments": map[string]interface{}{},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	json.Unmarshal([]byte(output), &resp)
	result := resp.Result.(map[string]interface{})
	if result["isError"] != true {
		t.Error("expected isError for empty queries")
	}
}

func TestToolCall_Prefetch(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	params := map[string]interface{}{
		"name":      "prefetch",
		"arguments": map[string]interface{}{"context_hint": "prefetch test", "limit": 3.0},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
}

func TestToolCall_PrefetchSuggest_MissingContext(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	params := map[string]interface{}{
		"name":      "prefetch_suggest",
		"arguments": map[string]interface{}{"limit": 5.0},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	json.Unmarshal([]byte(output), &resp)
	result := resp.Result.(map[string]interface{})
	if result["isError"] != true {
		t.Error("expected isError for missing context")
	}
}

// TestServer_Start_OneRequestThenEOF covers Start() stdio loop: one request then EOF
func TestServer_Start_OneRequestThenEOF(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PHLOEM_DATA_DIR", tmpDir)
	defer os.Unsetenv("PHLOEM_DATA_DIR")
	oldStderr := os.Stderr
	os.Stderr, _ = os.Open(os.DevNull)
	defer func() { os.Stderr = oldStderr }()

	initReq := JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "initialize"}
	line, _ := json.Marshal(initReq)

	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	go func() {
		w.Write(append(line, '\n'))
		w.Close()
	}()

	server, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer server.Stop()

	done := make(chan error, 1)
	go func() { done <- server.Start() }()

	select {
	case err := <-done:
		if err != nil && err.Error() != "EOF" && !strings.Contains(err.Error(), "read") {
			t.Errorf("Start after EOF: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Start timed out")
	}
}

func TestToolCall_Remember_WithCitations(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	tmpDir := t.TempDir()
	fpath := filepath.Join(tmpDir, "f.go")
	os.WriteFile(fpath, []byte("line1\nline2\nline3\n"), 0644)

	params := map[string]interface{}{
		"name": "remember",
		"arguments": map[string]interface{}{
			"content": "Memory with citation",
			"tags":    []interface{}{"code"},
			"context": "test",
			"citations": []interface{}{
				map[string]interface{}{
					"file_path":  fpath,
					"start_line": 1.0,
					"end_line":   2.0,
					"content":    "line1\nline2",
				},
			},
		},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	content := result["content"].([]interface{})
	text := content[0].(map[string]interface{})["text"].(string)
	if !strings.Contains(text, "citation") && !strings.Contains(text, "remembered") {
		t.Errorf("expected citation or remembered: %s", text)
	}
}

func TestToolCall_Compose_WithQueriesArray(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	server.store.Remember(ctx, "compose query one", nil, "")
	server.store.Remember(ctx, "compose query two", nil, "")

	params := map[string]interface{}{
		"name": "compose",
		"arguments": map[string]interface{}{
			"queries": []interface{}{"compose", "query"},
			"limit":   5.0,
		},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
}

func TestToolCall_ListMemories_WithLimit(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	server.store.Remember(ctx, "list mem one", nil, "")
	server.store.Remember(ctx, "list mem two", nil, "")

	params := map[string]interface{}{
		"name":      "list_memories",
		"arguments": map[string]interface{}{"limit": 1.0},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	content := result["content"].([]interface{})
	text := content[0].(map[string]interface{})["text"].(string)
	if !strings.Contains(text, "memories") && !strings.Contains(text, "id") {
		t.Errorf("expected memories in response: %s", text[:min(100, len(text))])
	}
}

func TestToolCall_Recall_WithTags(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	server.store.Remember(ctx, "recall tagged content", []string{"important"}, "")

	params := map[string]interface{}{
		"name": "recall",
		"arguments": map[string]interface{}{
			"query": "recall",
			"limit": 5.0,
			"tags":  []interface{}{"important"},
		},
	}
	paramsJSON, _ := json.Marshal(params)
	req := &JSONRPCRequest{JSONRPC: "2.0", ID: 1, Method: "tools/call", Params: paramsJSON}
	output := captureOutput(func() { server.handleRequest(req) })
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
}

// =============================================================================
// Truncate Helper Tests
// =============================================================================

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		max      int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly ten", 11, "exactly ten"},
		{"this is a long string", 10, "this is..."},
		{"", 10, ""},
		{"abc", 3, "abc"},
		{"abcd", 3, "..."},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.max)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, result, tt.expected)
		}
	}
}
