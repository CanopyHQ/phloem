// Package memory provides the local memory storage for Phloem
package memory

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/CanopyHQ/phloem/internal/memory/causal"
	_ "github.com/mattn/go-sqlite3"
)

// Citation links a memory to a specific location in code or documents
type Citation struct {
	ID         string    `json:"id"`
	MemoryID   string    `json:"memory_id"`
	FilePath   string    `json:"file_path"`            // Path to the file
	StartLine  int       `json:"start_line,omitempty"` // Starting line number
	EndLine    int       `json:"end_line,omitempty"`   // Ending line number
	CommitSHA  string    `json:"commit_sha,omitempty"` // Git commit SHA when citation was created
	Content    string    `json:"content,omitempty"`    // Snapshot of cited content for verification
	Confidence float64   `json:"confidence"`           // 0.0-1.0, decays over time
	VerifiedAt time.Time `json:"verified_at"`          // Last verification time
	CreatedAt  time.Time `json:"created_at"`
}

// Memory represents a stored memory
type Memory struct {
	ID           string     `json:"id"`
	Content      string     `json:"content"`
	Tags         []string   `json:"tags"`
	Context      string     `json:"context"`
	Scope        string     `json:"scope,omitempty"` // Repository scope (e.g., "github.com/owner/repo")
	Embedding    []float32  `json:"embedding,omitempty"`
	Citations    []Citation `json:"citations,omitempty"` // Linked citations
	Confidence   float64    `json:"confidence"`          // Aggregate confidence from citations
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Similarity   float64    `json:"similarity,omitempty"`    // Set during recall
	UtilityScore float64    `json:"utility_score,omitempty"` // 0.0-1.0 from memory critic; default 1.0
	Source       string     `json:"source,omitempty"`        // Attribution: "graft:name:author" or "user" or "sync"
}

// Edge represents a directed edge between memories (temporal, causal, or semantic)
type Edge struct {
	ID        string    `json:"id"`
	SourceID  string    `json:"source_id"`
	TargetID  string    `json:"target_id"`
	EdgeType  string    `json:"edge_type"` // temporal, causal, semantic
	Payload   string    `json:"payload,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Store provides local memory storage using SQLite
type Store struct {
	db       *sql.DB
	dataDir  string
	embedder Embedder

	// Vector index for fast KNN recall (nil if sqlite-vec unavailable)
	vecIdx *vecIndex
}

// GetDB returns the underlying SQL database handle
func (s *Store) GetDB() *sql.DB {
	return s.db
}

// NewStore creates a new memory store
func NewStore() (*Store, error) {
	// Determine data directory
	dataDir := os.Getenv("PHLOEM_DATA_DIR")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home dir: %w", err)
		}
		dataDir = filepath.Join(home, ".phloem")
	}

	// Create directory
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}

	// Open database
	dbPath := filepath.Join(dataDir, "memories.db")
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &Store{
		db:       db,
		dataDir:  dataDir,
		embedder: GetEmbedder(),
	}

	// Initialize schema
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	// Initialize sqlite-vec vector index for fast KNN recall
	store.vecIdx = newVecIndex(db, store.embedder.Dimensions())
	if store.vecIdx.available {
		if n, err := store.vecIdx.Backfill(db); err == nil && n > 0 {
			fmt.Fprintf(os.Stderr, "🔍 Backfilled %d memories into vec index\n", n)
		}
	}

	fmt.Fprintf(os.Stderr, "📁 Memory store: %s\n", dbPath)
	return store, nil
}

// initSchema creates the database tables
func (s *Store) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS memories (
		id TEXT PRIMARY KEY,
		content TEXT NOT NULL,
		content_hash TEXT,
		tags TEXT,
		context TEXT,
		embedding BLOB,
		scope TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		utility_score REAL DEFAULT 1.0
	);

	CREATE INDEX IF NOT EXISTS idx_memories_created_at ON memories(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_memories_tags ON memories(tags);

	CREATE TABLE IF NOT EXISTS memory_tags (
		memory_id TEXT,
		tag TEXT,
		FOREIGN KEY (memory_id) REFERENCES memories(id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_memory_tags_tag ON memory_tags(tag);

	CREATE TABLE IF NOT EXISTS citations (
		id TEXT PRIMARY KEY,
		memory_id TEXT NOT NULL,
		file_path TEXT NOT NULL,
		start_line INTEGER,
		end_line INTEGER,
		commit_sha TEXT,
		content TEXT,
		confidence REAL DEFAULT 1.0,
		verified_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (memory_id) REFERENCES memories(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_citations_memory_id ON citations(memory_id);
	CREATE INDEX IF NOT EXISTS idx_citations_file_path ON citations(file_path);

	CREATE TABLE IF NOT EXISTS memory_edges (
		id TEXT PRIMARY KEY,
		source_id TEXT NOT NULL,
		target_id TEXT,
		edge_type TEXT NOT NULL,
		payload TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (source_id) REFERENCES memories(id) ON DELETE CASCADE,
		FOREIGN KEY (target_id) REFERENCES memories(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_memory_edges_source_type ON memory_edges(source_id, edge_type);
	CREATE INDEX IF NOT EXISTS idx_memory_edges_target_type ON memory_edges(target_id, edge_type);
	`
	_, err := s.db.Exec(schema)
	if err != nil {
		return err
	}

	// Migrate: Add content_hash column if it doesn't exist
	_, _ = s.db.Exec(`ALTER TABLE memories ADD COLUMN content_hash TEXT`)
	// Create index (ignore if exists)
	_, _ = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_memories_content_hash ON memories(content_hash)`)

	// Migrate: Add utility_score for memory critic (Stage 3); default 1.0 = full weight in recall
	_, _ = s.db.Exec(`ALTER TABLE memories ADD COLUMN utility_score REAL DEFAULT 1.0`)

	// Migrate: Add scope support for repo-scoped memories
	_, _ = s.db.Exec(`ALTER TABLE memories ADD COLUMN scope TEXT`)
	_, _ = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_memories_scope ON memories(scope)`)

	// Create scopes table
	_, _ = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS scopes (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			metadata TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	_, _ = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_scopes_name ON scopes(name)`)
	_, _ = s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_scopes_type ON scopes(type)`)

	// Migrate: Add source column for graft attribution tracking
	_, _ = s.db.Exec(`ALTER TABLE memories ADD COLUMN source TEXT DEFAULT ''`)

	return nil
}

// AddEdge adds a directed edge between memories (temporal, causal, or semantic)
func (s *Store) AddEdge(ctx context.Context, sourceID, targetID, edgeType, payload string) error {
	if sourceID == "" || edgeType == "" {
		return nil
	}
	id := generateID()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO memory_edges (id, source_id, target_id, edge_type, payload, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, id, sourceID, targetID, edgeType, payload, time.Now())
	return err
}

// GetEdgesFrom returns edges originating from the given memory, optionally filtered by type
func (s *Store) GetEdgesFrom(ctx context.Context, memoryID string, edgeType string) ([]Edge, error) {
	sqlQuery := `SELECT id, source_id, target_id, edge_type, payload, created_at FROM memory_edges WHERE source_id = ?`
	args := []interface{}{memoryID}
	if edgeType != "" {
		sqlQuery += ` AND edge_type = ?`
		args = append(args, edgeType)
	}
	sqlQuery += ` ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var edges []Edge
	for rows.Next() {
		var e Edge
		var targetIDVal, payloadVal sql.NullString
		if err := rows.Scan(&e.ID, &e.SourceID, &targetIDVal, &e.EdgeType, &payloadVal, &e.CreatedAt); err != nil {
			continue
		}
		if targetIDVal.Valid {
			e.TargetID = targetIDVal.String
		}
		if payloadVal.Valid {
			e.Payload = payloadVal.String
		}
		edges = append(edges, e)
	}
	return edges, nil
}

// GetEdgesTo returns edges pointing to the given memory, optionally filtered by type
func (s *Store) GetEdgesTo(ctx context.Context, memoryID string, edgeType string) ([]Edge, error) {
	sqlQuery := `SELECT id, source_id, target_id, edge_type, payload, created_at FROM memory_edges WHERE target_id = ?`
	args := []interface{}{memoryID}
	if edgeType != "" {
		sqlQuery += ` AND edge_type = ?`
		args = append(args, edgeType)
	}
	sqlQuery += ` ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var edges []Edge
	for rows.Next() {
		var e Edge
		var targetIDVal, payloadVal sql.NullString
		if err := rows.Scan(&e.ID, &e.SourceID, &targetIDVal, &e.EdgeType, &payloadVal, &e.CreatedAt); err != nil {
			continue
		}
		if targetIDVal.Valid {
			e.TargetID = targetIDVal.String
		}
		if payloadVal.Valid {
			e.Payload = payloadVal.String
		}
		edges = append(edges, e)
	}
	return edges, nil
}

// GetIdentityProfile fetches memories tagged with 'identity:profile'
func (s *Store) GetIdentityProfile(ctx context.Context) ([]*Memory, error) {
	return s.List(ctx, 100, []string{"identity:profile"})
}

// GetMemoryByID returns a single memory by ID, or nil if not found.
func (s *Store) GetMemoryByID(ctx context.Context, id string) (*Memory, error) {
	if id == "" {
		return nil, nil
	}
	row := s.db.QueryRowContext(ctx, `
		SELECT id, content, tags, context, scope, embedding, created_at, updated_at, COALESCE(utility_score, 1.0)
		FROM memories WHERE id = ?
	`, id)
	// Use a single row scanner; scanMemory expects *sql.Rows, so we need a small adapter or duplicate scan logic.
	var mem Memory
	var tagsJSON, embeddingJSON string
	var contextNull, scopeNull sql.NullString
	var utilityNull sql.NullFloat64
	err := row.Scan(&mem.ID, &mem.Content, &tagsJSON, &contextNull, &scopeNull, &embeddingJSON, &mem.CreatedAt, &mem.UpdatedAt, &utilityNull)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if contextNull.Valid {
		mem.Context = contextNull.String
	}
	if scopeNull.Valid {
		mem.Scope = scopeNull.String
	}
	if utilityNull.Valid {
		mem.UtilityScore = utilityNull.Float64
	} else {
		mem.UtilityScore = 1.0
	}
	_ = json.Unmarshal([]byte(tagsJSON), &mem.Tags)
	_ = json.Unmarshal([]byte(embeddingJSON), &mem.Embedding)
	return &mem, nil
}

// CausalNeighbors returns memories that are directly connected to the given memory by causal edges (one hop, both directions).
func (s *Store) CausalNeighbors(ctx context.Context, memoryID string) ([]*Memory, error) {
	fromEdges, _ := s.GetEdgesFrom(ctx, memoryID, "causal")
	toEdges, _ := s.GetEdgesTo(ctx, memoryID, "causal")
	seen := map[string]bool{memoryID: true}
	var ids []string
	for _, e := range fromEdges {
		if e.TargetID != "" && !seen[e.TargetID] {
			seen[e.TargetID] = true
			ids = append(ids, e.TargetID)
		}
	}
	for _, e := range toEdges {
		if e.SourceID != "" && !seen[e.SourceID] {
			seen[e.SourceID] = true
			ids = append(ids, e.SourceID)
		}
	}
	var out []*Memory
	for _, id := range ids {
		mem, err := s.GetMemoryByID(ctx, id)
		if err != nil || mem == nil {
			continue
		}
		out = append(out, mem)
	}
	return out, nil
}

// AffectedIfChanged returns all memory IDs that would be "affected" if the given memory changed (transitive descendants via causal edges).
// Follows causal edges from memoryID downstream (source -> target). Used for "what breaks if I change X?"
func (s *Store) AffectedIfChanged(ctx context.Context, memoryID string) ([]string, error) {
	visited := map[string]bool{}
	var queue []string
	queue = append(queue, memoryID)
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if visited[cur] {
			continue
		}
		visited[cur] = true
		edges, err := s.GetEdgesFrom(ctx, cur, "causal")
		if err != nil {
			continue
		}
		for _, e := range edges {
			if e.TargetID != "" && !visited[e.TargetID] {
				queue = append(queue, e.TargetID)
			}
		}
	}
	// Exclude the seed memory itself
	delete(visited, memoryID)
	var out []string
	for id := range visited {
		out = append(out, id)
	}
	sort.Strings(out)
	return out, nil
}

// GetPreviousMemoryID returns the ID of the most recent memory before the given time (for temporal edges)
func (s *Store) GetPreviousMemoryID(ctx context.Context, before time.Time) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `
		SELECT id FROM memories WHERE created_at < ? ORDER BY created_at DESC LIMIT 1
	`, before).Scan(&id)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return id, err
}

// RunCausalExtractionAsync runs causal relation extraction on the memory in the background
// and adds causal edges when related memories are found. Does not block.
func (s *Store) RunCausalExtractionAsync(mem *Memory) {
	if mem == nil {
		return
	}
	go s.runCausalExtraction(mem)
}

func (s *Store) runCausalExtraction(mem *Memory) {
	ctx := context.Background()
	rels := causal.Extract(mem.Content)
	for _, r := range rels {
		if r.Phrase == "" {
			continue
		}
		results, err := s.Recall(ctx, r.Phrase, 1, nil)
		if err != nil || len(results) == 0 {
			continue
		}
		target := results[0]
		if target.ID == mem.ID {
			continue
		}
		_ = s.AddEdge(ctx, mem.ID, target.ID, "causal", r.Reason)
	}
}

// SetMemoryUtility sets the utility score for a memory (0.0-1.0). Used by memory critic; low score deprioritizes in recall.
func (s *Store) SetMemoryUtility(ctx context.Context, memoryID string, score float64) error {
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	_, err := s.db.ExecContext(ctx, `UPDATE memories SET utility_score = ? WHERE id = ?`, score, memoryID)
	return err
}

// RunMemoryDreams runs an offline pass: link similar memories with semantic edges.
// Call periodically (e.g. nightly). Lists recent memories, for each finds similar via recall, adds semantic edges.
// Does not run DecayCitations; call that separately if desired.
func (s *Store) RunMemoryDreams(ctx context.Context, recentLimit, linksPerMemory int) (edgesAdded int, err error) {
	if recentLimit <= 0 {
		recentLimit = 30
	}
	if linksPerMemory <= 0 {
		linksPerMemory = 3
	}
	memories, err := s.List(ctx, recentLimit, nil)
	if err != nil || len(memories) == 0 {
		return 0, err
	}
	for _, mem := range memories {
		if len(mem.Content) < 10 {
			continue
		}
		similar, err := s.Recall(ctx, mem.Content, linksPerMemory+1, nil) // +1 to account for self
		if err != nil {
			continue
		}
		existing, _ := s.GetEdgesFrom(ctx, mem.ID, "semantic")
		haveTarget := make(map[string]bool)
		for _, e := range existing {
			if e.TargetID != "" {
				haveTarget[e.TargetID] = true
			}
		}
		for _, m := range similar {
			if m == nil || m.ID == mem.ID || haveTarget[m.ID] {
				continue
			}
			if err := s.AddEdge(ctx, mem.ID, m.ID, "semantic", ""); err == nil {
				edgesAdded++
				haveTarget[m.ID] = true
			}
		}
	}
	return edgesAdded, nil
}

// RunMemoryCritic updates utility scores from citation confidence (rules-based v1). Call periodically (e.g. after DecayCitations).
// Utility = 0.5 + 0.5*confidence so range is 0.5-1.0; memories with no citations stay at 1.0.
func (s *Store) RunMemoryCritic(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM memories`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		conf, err := s.GetMemoryConfidence(ctx, id)
		if err != nil {
			continue
		}
		utility := 0.5 + 0.5*conf
		_ = s.SetMemoryUtility(ctx, id, utility)
	}
	return nil
}

// NightlyCurationResult summarizes the outcome of RunNightlyCuration.
type NightlyCurationResult struct {
	DecayedCitations int
	DreamsEdgesAdded int
	Error            string
}

// RunNightlyCuration runs the full offline curation pass (Stage 3): decay citations, update utility from confidence, link similar memories.
// Call periodically (e.g. nightly). Does not delete memories; only adjusts weights and adds semantic edges.
func (s *Store) RunNightlyCuration(ctx context.Context) (NightlyCurationResult, error) {
	var result NightlyCurationResult
	decayed, err := s.DecayCitations(ctx)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}
	result.DecayedCitations = decayed
	if err := s.RunMemoryCritic(ctx); err != nil {
		result.Error = err.Error()
		return result, err
	}
	edgesAdded, err := s.RunMemoryDreams(ctx, 30, 3)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}
	result.DreamsEdgesAdded = edgesAdded
	return result, nil
}

// contentHash calculates SHA256 hash of content for deduplication
func contentHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}

// Add inserts a memory directly into the store (used for imports)
func (s *Store) Add(ctx context.Context, m Memory) error {
	// Generate ID if missing
	if m.ID == "" {
		m.ID = generateID()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now()
	}

	// Generate hash for deduplication
	hash := sha256.Sum256([]byte(m.Content))
	contentHash := hex.EncodeToString(hash[:])

	// Check for existing memory with same content hash AND scope
	var existingID string
	var existingTagsJSON string
	query := `SELECT id, tags FROM memories WHERE content_hash = ? AND (scope = ? OR (scope IS NULL AND ? = ''))`
	err := s.db.QueryRowContext(ctx, query, contentHash, m.Scope, m.Scope).Scan(&existingID, &existingTagsJSON)
	if err == nil {
		// Already exists, skip or update? For now, skip to avoid duplicates
		return nil
	}

	// Ensure embedding
	if len(m.Embedding) == 0 {
		embedding, err := s.embedder.Embed(m.Content)
		if err == nil {
			m.Embedding = embedding
		} else {
			// Fallback to empty if embedding fails
			m.Embedding = make([]float32, s.embedder.Dimensions())
		}
	}

	tagsJSON, _ := json.Marshal(m.Tags)
	embeddingJSON, _ := json.Marshal(m.Embedding)

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO memories (id, content, content_hash, tags, context, embedding, created_at, updated_at, source)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, m.ID, m.Content, contentHash, string(tagsJSON), m.Context, embeddingJSON, m.CreatedAt, m.UpdatedAt, m.Source)

	if err != nil {
		return fmt.Errorf("failed to insert memory: %w", err)
	}

	// Insert into vec index
	if s.vecIdx != nil {
		s.vecIdx.Insert(m.ID, m.Embedding)
	}

	// Insert tags
	for _, tag := range m.Tags {
		s.db.ExecContext(ctx, `INSERT OR IGNORE INTO memory_tags (memory_id, tag) VALUES (?, ?)`, m.ID, tag)
	}

	return nil
}

// Remember stores a new memory, checking for duplicates by content hash
func (s *Store) Remember(ctx context.Context, content string, tags []string, memContext string) (*Memory, error) {
	return s.RememberWithScope(ctx, content, tags, memContext, "")
}

// RememberWithScope stores a new memory with a scope identifier
func (s *Store) RememberWithScope(ctx context.Context, content string, tags []string, memContext string, scope string) (*Memory, error) {
	// Calculate content hash for deduplication
	hash := contentHash(content)

	// Check for existing memory with same content hash AND scope
	var existingID string
	var existingTagsJSON string
	var query string
	var args []interface{}

	if scope == "" {
		query = `SELECT id, tags FROM memories WHERE content_hash = ? AND (scope IS NULL OR scope = '')`
		args = []interface{}{hash}
	} else {
		query = `SELECT id, tags FROM memories WHERE content_hash = ? AND scope = ?`
		args = []interface{}{hash, scope}
	}

	err := s.db.QueryRowContext(ctx, query, args...).Scan(&existingID, &existingTagsJSON)

	if err == nil {
		// Duplicate found - merge tags and update
		var existingTags []string
		if existingTagsJSON != "" {
			json.Unmarshal([]byte(existingTagsJSON), &existingTags)
		}

		// Merge tags (deduplicate)
		tagMap := make(map[string]bool)
		for _, tag := range existingTags {
			tagMap[tag] = true
		}
		for _, tag := range tags {
			tagMap[tag] = true
		}

		mergedTags := make([]string, 0, len(tagMap))
		for tag := range tagMap {
			mergedTags = append(mergedTags, tag)
		}
		sort.Strings(mergedTags)

		// Update existing memory with merged tags
		mergedTagsJSON, _ := json.Marshal(mergedTags)
		now := time.Now()
		_, err = s.db.ExecContext(ctx, `
			UPDATE memories SET tags = ?, updated_at = ? WHERE id = ?
		`, string(mergedTagsJSON), now, existingID)

		if err != nil {
			return nil, fmt.Errorf("failed to update duplicate memory: %w", err)
		}

		// Update memory_tags table
		s.db.ExecContext(ctx, `DELETE FROM memory_tags WHERE memory_id = ?`, existingID)
		for _, tag := range mergedTags {
			s.db.ExecContext(ctx, `INSERT INTO memory_tags (memory_id, tag) VALUES (?, ?)`, existingID, tag)
		}

		// Return existing memory
		var existingMemory Memory
		var embeddingJSON []byte
		var utilityNull sql.NullFloat64
		err = s.db.QueryRowContext(ctx, `
			SELECT id, content, tags, context, embedding, created_at, updated_at, COALESCE(utility_score, 1.0)
			FROM memories WHERE id = ?
		`, existingID).Scan(&existingMemory.ID, &existingMemory.Content, &existingTagsJSON,
			&existingMemory.Context, &embeddingJSON, &existingMemory.CreatedAt, &existingMemory.UpdatedAt, &utilityNull)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve updated memory: %w", err)
		}
		if utilityNull.Valid {
			existingMemory.UtilityScore = utilityNull.Float64
		} else {
			existingMemory.UtilityScore = 1.0
		}

		json.Unmarshal([]byte(existingTagsJSON), &existingMemory.Tags)
		if len(embeddingJSON) > 0 {
			json.Unmarshal(embeddingJSON, &existingMemory.Embedding)
		}

		return &existingMemory, nil
	}

	// No duplicate - create new memory
	id := generateID()
	now := time.Now()

	// Generate embedding for semantic search using the configured embedder
	embedding, err := s.embedder.Embed(content)
	if err != nil {
		// Fall back to empty embedding if embedding fails
		fmt.Fprintf(os.Stderr, "⚠️  Embedding failed: %v\n", err)
		embedding = make([]float32, s.embedder.Dimensions())
	}

	tagsJSON, _ := json.Marshal(tags)
	embeddingJSON, _ := json.Marshal(embedding)

	_, dbErr := s.db.ExecContext(ctx, `
		INSERT INTO memories (id, content, content_hash, tags, context, scope, embedding, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, content, hash, string(tagsJSON), memContext, scope, embeddingJSON, now, now)

	if dbErr != nil {
		return nil, fmt.Errorf("failed to store memory: %w", dbErr)
	}

	// Store individual tags
	for _, tag := range tags {
		s.db.ExecContext(ctx, `INSERT INTO memory_tags (memory_id, tag) VALUES (?, ?)`, id, tag)
	}

	mem := &Memory{
		ID:        id,
		Content:   content,
		Tags:      tags,
		Context:   memContext,
		Scope:     scope,
		Embedding: embedding,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Insert into vec index for fast KNN recall
	if s.vecIdx != nil {
		s.vecIdx.Insert(id, embedding)
	}

	// Temporal edge: link from previous memory (by created_at) to this one
	if prevID, err := s.GetPreviousMemoryID(ctx, now); err == nil && prevID != "" {
		_ = s.AddEdge(ctx, prevID, id, "temporal", "")
	}

	// Causal extraction: async so it does not block MCP
	s.RunCausalExtractionAsync(mem)

	return mem, nil
}

// Recall finds memories similar to the query
// Optimized with early filtering for better performance on large datasets
func (s *Store) Recall(ctx context.Context, query string, limit int, filterTags []string) ([]*Memory, error) {
	return s.RecallWithScope(ctx, query, limit, filterTags, "")
}

// RecallWithScope finds memories similar to the query, optionally filtered by scope
func (s *Store) RecallWithScope(ctx context.Context, query string, limit int, filterTags []string, scope string) ([]*Memory, error) {
	// Generate query embedding using the configured embedder
	queryEmbedding, err := s.embedder.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Fast path: use sqlite-vec KNN index when available
	if s.vecIdx != nil && s.vecIdx.available {
		results, err := s.recallWithVecIndex(ctx, queryEmbedding, limit, filterTags, scope)
		if err == nil && len(results) > 0 {
			return results, nil
		}
		// Fall through to linear scan on error or empty results
	}

	// Performance optimization: Use recency-based pre-filtering for large datasets
	// This reduces the number of embeddings we need to compare
	// Note: Only use hybrid recall if no tag filtering (RecallWithRecencyBoost doesn't support tags yet)
	var count int
	err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM memories`).Scan(&count)
	if err == nil && count > 5000 && len(filterTags) == 0 {
		// For large datasets without tag filtering, use optimized recall with recency boost
		return s.RecallWithRecencyBoost(ctx, query, limit, RecallOptions{
			SemanticWeight:       0.7,
			RecencyWeight:        0.3,
			RecencyHalfLifeHours: 168,                                  // 1 week
			Since:                time.Now().Add(-90 * 24 * time.Hour), // Last 90 days
		})
	}

	// Linear scan fallback
	return s.recallLinearScan(ctx, queryEmbedding, limit, filterTags, scope)
}

// recallWithVecIndex uses the sqlite-vec KNN index for fast recall.
// Overfetches candidates then applies tag/scope/utility filtering in Go.
func (s *Store) recallWithVecIndex(ctx context.Context, queryEmbedding []float32, limit int, filterTags []string, scope string) ([]*Memory, error) {
	// Overfetch to allow for filtering and utility re-ranking
	overfetchLimit := limit * 3
	if len(filterTags) > 0 || scope != "" {
		overfetchLimit = limit * 5
	}
	if overfetchLimit < 20 {
		overfetchLimit = 20
	}

	results, err := s.vecIdx.Search(queryEmbedding, overfetchLimit)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}

	// Build distance lookup and ID list for batch fetch
	distanceMap := make(map[string]float64, len(results))
	placeholders := make([]string, len(results))
	args := make([]interface{}, len(results))
	for i, r := range results {
		distanceMap[r.MemoryID] = r.Distance
		placeholders[i] = "?"
		args[i] = r.MemoryID
	}

	// Batch-fetch full memory data
	sqlQuery := `SELECT id, content, tags, context, scope, embedding, created_at, updated_at, COALESCE(utility_score, 1.0)
		FROM memories WHERE id IN (` + strings.Join(placeholders, ",") + `)`

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []*Memory
	for rows.Next() {
		mem, err := s.scanMemory(rows)
		if err != nil {
			continue
		}

		// Apply scope filter
		if scope != "" && mem.Scope != scope {
			continue
		}

		// Apply tag filter
		if len(filterTags) > 0 && !memHasAnyTag(mem, filterTags) {
			continue
		}

		// Convert cosine distance to similarity, apply utility score
		distance := distanceMap[mem.ID]
		similarity := 1.0 - distance
		if mem.UtilityScore <= 0 {
			mem.UtilityScore = 0.5
		}
		mem.Similarity = similarity * mem.UtilityScore

		memories = append(memories, mem)
	}

	// Sort by similarity (utility may change vec0's distance order)
	sort.Slice(memories, func(i, j int) bool {
		return memories[i].Similarity > memories[j].Similarity
	})

	if len(memories) > limit {
		memories = memories[:limit]
	}

	return memories, nil
}

// recallLinearScan is the original brute-force recall path (fallback when vec index is unavailable).
func (s *Store) recallLinearScan(ctx context.Context, queryEmbedding []float32, limit int, filterTags []string, scope string) ([]*Memory, error) {
	// Build query with optional tag and scope filtering
	sqlQuery := `SELECT id, content, tags, context, scope, embedding, created_at, updated_at, COALESCE(utility_score, 1.0) FROM memories`
	args := []interface{}{}
	whereConditions := []string{}

	if scope != "" {
		whereConditions = append(whereConditions, "scope = ?")
		args = append(args, scope)
	}

	if len(filterTags) > 0 {
		placeholders := make([]string, len(filterTags))
		for i, tag := range filterTags {
			placeholders[i] = "?"
			args = append(args, tag)
		}
		whereConditions = append(whereConditions, `id IN (SELECT memory_id FROM memory_tags WHERE tag IN (`+strings.Join(placeholders, ",")+`))`)
	}

	if len(whereConditions) > 0 {
		sqlQuery += ` WHERE ` + strings.Join(whereConditions, " AND ")
	}

	// Add ORDER BY to leverage index for better performance
	sqlQuery += ` ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query memories: %w", err)
	}
	defer rows.Close()

	var memories []*Memory
	for rows.Next() {
		mem, err := s.scanMemory(rows)
		if err != nil {
			continue
		}
		if mem.UtilityScore <= 0 {
			mem.UtilityScore = 0.5 // avoid zero so we don't drop; critic can demote
		}
		// Calculate similarity
		mem.Similarity = cosineSimilarity(queryEmbedding, mem.Embedding) * mem.UtilityScore
		memories = append(memories, mem)
	}

	// Sort by similarity
	sort.Slice(memories, func(i, j int) bool {
		return memories[i].Similarity > memories[j].Similarity
	})

	// Limit results
	if len(memories) > limit {
		memories = memories[:limit]
	}

	return memories, nil
}

// memHasAnyTag returns true if the memory has at least one of the given tags.
func memHasAnyTag(mem *Memory, tags []string) bool {
	tagSet := make(map[string]bool, len(mem.Tags))
	for _, t := range mem.Tags {
		tagSet[t] = true
	}
	for _, t := range tags {
		if tagSet[t] {
			return true
		}
	}
	return false
}

// RecallWithRecencyBoost finds memories using blended scoring:
// FinalScore = (semantic × semanticWeight) + (recency × recencyWeight) + (importance × importanceWeight)
// This ensures recent memories surface even if semantic match is imperfect.
func (s *Store) RecallWithRecencyBoost(ctx context.Context, query string, limit int, options RecallOptions) ([]*Memory, error) {
	// Generate query embedding
	queryEmbedding, err := s.embedder.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Fast path: use vec index for semantic candidates, then blend with recency
	if s.vecIdx != nil && s.vecIdx.available {
		return s.recallWithRecencyBoostVec(ctx, queryEmbedding, limit, options)
	}

	// Fallback: full scan
	return s.recallWithRecencyBoostLinear(ctx, queryEmbedding, limit, options)
}

// recallWithRecencyBoostVec uses the vec index for semantic candidates, then applies blended scoring.
func (s *Store) recallWithRecencyBoostVec(ctx context.Context, queryEmbedding []float32, limit int, options RecallOptions) ([]*Memory, error) {
	// Get top semantic candidates from vec index (overfetch for blending)
	candidateLimit := limit * 5
	if candidateLimit < 50 {
		candidateLimit = 50
	}

	vecResults, err := s.vecIdx.Search(queryEmbedding, candidateLimit)
	if err != nil || len(vecResults) == 0 {
		// Fall back to linear scan
		return s.recallWithRecencyBoostLinear(ctx, queryEmbedding, limit, options)
	}

	// Build distance lookup and fetch candidate memories
	distanceMap := make(map[string]float64, len(vecResults))
	placeholders := make([]string, len(vecResults))
	args := make([]interface{}, len(vecResults))
	for i, r := range vecResults {
		distanceMap[r.MemoryID] = r.Distance
		placeholders[i] = "?"
		args[i] = r.MemoryID
	}

	sqlQuery := `SELECT id, content, tags, context, scope, embedding, created_at, updated_at, COALESCE(utility_score, 1.0)
		FROM memories WHERE id IN (` + strings.Join(placeholders, ",") + `)`

	// Apply time window filter if specified
	if !options.Since.IsZero() {
		sqlQuery += ` AND created_at >= ?`
		args = append(args, options.Since)
	}

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	now := time.Now()
	var memories []*Memory

	for rows.Next() {
		mem, err := s.scanMemory(rows)
		if err != nil {
			continue
		}
		if mem.UtilityScore <= 0 {
			mem.UtilityScore = 0.5
		}

		// Semantic score from vec distance
		semantic := 1.0 - distanceMap[mem.ID]

		blended := s.computeBlendedScore(ctx, mem, semantic, now, options)
		mem.Similarity = blended * mem.UtilityScore

		memories = append(memories, mem)
	}

	// Sort by blended score
	sort.Slice(memories, func(i, j int) bool {
		return memories[i].Similarity > memories[j].Similarity
	})

	if len(memories) > limit {
		memories = memories[:limit]
	}

	return memories, nil
}

// recallWithRecencyBoostLinear is the original full-scan blended recall.
func (s *Store) recallWithRecencyBoostLinear(ctx context.Context, queryEmbedding []float32, limit int, options RecallOptions) ([]*Memory, error) {
	sqlQuery := `SELECT id, content, tags, context, scope, embedding, created_at, updated_at, COALESCE(utility_score, 1.0) FROM memories`
	args := []interface{}{}

	// Optional time window filter for efficiency at scale
	if !options.Since.IsZero() {
		sqlQuery += ` WHERE created_at >= ?`
		args = append(args, options.Since)
	}

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query memories: %w", err)
	}
	defer rows.Close()

	now := time.Now()
	var memories []*Memory

	for rows.Next() {
		mem, err := s.scanMemory(rows)
		if err != nil {
			continue
		}
		if mem.UtilityScore <= 0 {
			mem.UtilityScore = 0.5
		}
		// Calculate semantic similarity
		semantic := cosineSimilarity(queryEmbedding, mem.Embedding)

		blended := s.computeBlendedScore(ctx, mem, semantic, now, options)
		mem.Similarity = blended * mem.UtilityScore

		memories = append(memories, mem)
	}

	// Sort by blended score
	sort.Slice(memories, func(i, j int) bool {
		return memories[i].Similarity > memories[j].Similarity
	})

	// Limit results
	if len(memories) > limit {
		memories = memories[:limit]
	}

	return memories, nil
}

// computeBlendedScore calculates the blended recall score for a memory.
func (s *Store) computeBlendedScore(ctx context.Context, mem *Memory, semantic float64, now time.Time, options RecallOptions) float64 {
	// Calculate recency score: exponential decay with configurable half-life
	halfLife := options.RecencyHalfLifeHours
	if halfLife <= 0 {
		halfLife = 168 // 1 week default
	}
	hoursAgo := now.Sub(mem.CreatedAt).Hours()
	recency := math.Exp(-hoursAgo * math.Ln2 / halfLife)

	// Calculate importance boost based on tags
	importance := 0.0
	for _, tag := range mem.Tags {
		switch tag {
		case "critical", "milestone", "founding", "permanent", "promise":
			importance = 1.0
		case "decision", "architecture":
			if importance < 0.5 {
				importance = 0.5
			}
		}
	}

	// Get citation confidence for this memory
	confidence, err := s.GetMemoryConfidence(ctx, mem.ID)
	if err != nil {
		confidence = 1.0
	}

	// Apply weights (default: 50% semantic, 25% recency, 10% importance, 15% confidence)
	semanticWeight := options.SemanticWeight
	recencyWeight := options.RecencyWeight
	importanceWeight := options.ImportanceWeight
	confidenceWeight := options.ConfidenceWeight
	if semanticWeight <= 0 && recencyWeight <= 0 && importanceWeight <= 0 && confidenceWeight <= 0 {
		semanticWeight = 0.5
		recencyWeight = 0.25
		importanceWeight = 0.1
		confidenceWeight = 0.15
	}

	// Normalize weights
	totalWeight := semanticWeight + recencyWeight + importanceWeight + confidenceWeight
	if totalWeight > 0 {
		semanticWeight /= totalWeight
		recencyWeight /= totalWeight
		importanceWeight /= totalWeight
		confidenceWeight /= totalWeight
	}

	mem.Confidence = confidence
	return (semantic * semanticWeight) + (recency * recencyWeight) + (importance * importanceWeight) + (confidence * confidenceWeight)
}

// RecallOptions configures the blended recall algorithm
type RecallOptions struct {
	// Weight for semantic similarity (default 0.5)
	SemanticWeight float64
	// Weight for recency score (default 0.25)
	RecencyWeight float64
	// Weight for importance tags (default 0.1)
	ImportanceWeight float64
	// Weight for citation confidence (default 0.15)
	ConfidenceWeight float64
	// Half-life for recency decay in hours (default 168 = 1 week)
	RecencyHalfLifeHours float64
	// Only consider memories since this time (for efficiency at scale)
	Since time.Time
}

// GetRecentImportant returns recent memories with important tags, guaranteed to surface
// regardless of semantic similarity. Used by session_context for guaranteed slots.
func (s *Store) GetRecentImportant(ctx context.Context, maxAge time.Duration, limit int) ([]*Memory, error) {
	cutoff := time.Now().Add(-maxAge)

	sqlQuery := `
		SELECT DISTINCT m.id, m.content, m.tags, m.context, m.scope, m.embedding, m.created_at, m.updated_at, COALESCE(m.utility_score, 1.0)
		FROM memories m
		JOIN memory_tags mt ON m.id = mt.memory_id
		WHERE m.created_at >= ? 
		AND mt.tag IN ('critical', 'milestone', 'founding', 'permanent', 'promise', 'decision')
		ORDER BY m.created_at DESC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, sqlQuery, cutoff, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query important memories: %w", err)
	}
	defer rows.Close()

	var memories []*Memory
	for rows.Next() {
		mem, err := s.scanMemory(rows)
		if err != nil {
			continue
		}
		memories = append(memories, mem)
	}

	return memories, nil
}

// Forget deletes a memory
func (s *Store) Forget(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM memories WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("memory not found: %s", id)
	}

	// Also delete tags and vec index entry
	s.db.ExecContext(ctx, `DELETE FROM memory_tags WHERE memory_id = ?`, id)
	if s.vecIdx != nil {
		s.vecIdx.Delete(id)
	}

	return nil
}

// List returns recent memories
func (s *Store) List(ctx context.Context, limit int, filterTags []string) ([]*Memory, error) {
	sqlQuery := `SELECT id, content, tags, context, scope, embedding, created_at, updated_at, COALESCE(utility_score, 1.0) FROM memories`
	args := []interface{}{}

	if len(filterTags) > 0 {
		placeholders := make([]string, len(filterTags))
		for i, tag := range filterTags {
			placeholders[i] = "?"
			args = append(args, tag)
		}
		sqlQuery += ` WHERE id IN (SELECT memory_id FROM memory_tags WHERE tag IN (` + strings.Join(placeholders, ",") + `))`
	}

	sqlQuery += ` ORDER BY created_at DESC`
	if limit > 0 {
		sqlQuery += ` LIMIT ?`
		args = append(args, limit)
	}

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list memories: %w", err)
	}
	defer rows.Close()

	var memories []*Memory
	for rows.Next() {
		mem, err := s.scanMemory(rows)
		if err != nil {
			continue
		}

		memories = append(memories, mem)
	}

	return memories, nil
}

// Count returns the total number of memories
func (s *Store) Count(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM memories`).Scan(&count)
	return count, err
}

// Size returns the database file size as a human-readable string
func (s *Store) Size() (string, error) {
	dbPath := filepath.Join(s.dataDir, "memories.db")
	info, err := os.Stat(dbPath)
	if err != nil {
		return "unknown", err
	}

	size := info.Size()
	if size < 1024 {
		return fmt.Sprintf("%d B", size), nil
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024), nil
	} else {
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024)), nil
	}
}

// LastActivity returns the timestamp of the most recent memory
func (s *Store) LastActivity(ctx context.Context) (time.Time, error) {
	var lastActivityStr sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT MAX(updated_at) FROM memories`).Scan(&lastActivityStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	if !lastActivityStr.Valid || lastActivityStr.String == "" {
		return time.Time{}, nil
	}
	// Parse SQLite datetime format
	lastActivity, err := time.Parse("2006-01-02 15:04:05.999999999-07:00", lastActivityStr.String)
	if err != nil {
		// Try alternative formats
		lastActivity, err = time.Parse("2006-01-02T15:04:05Z", lastActivityStr.String)
		if err != nil {
			lastActivity, err = time.Parse(time.RFC3339Nano, lastActivityStr.String)
		}
	}
	return lastActivity, err
}

// Close closes the database
func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) scanMemory(rows *sql.Rows) (*Memory, error) {
	var mem Memory
	var tagsJSON, embeddingJSON string
	var contextNull, scopeNull sql.NullString
	var utilityNull sql.NullFloat64

	err := rows.Scan(&mem.ID, &mem.Content, &tagsJSON, &contextNull, &scopeNull, &embeddingJSON, &mem.CreatedAt, &mem.UpdatedAt, &utilityNull)
	if err != nil {
		return nil, err
	}

	if contextNull.Valid {
		mem.Context = contextNull.String
	}
	if scopeNull.Valid {
		mem.Scope = scopeNull.String
	}
	if utilityNull.Valid {
		mem.UtilityScore = utilityNull.Float64
	} else {
		mem.UtilityScore = 1.0
	}

	json.Unmarshal([]byte(tagsJSON), &mem.Tags)
	json.Unmarshal([]byte(embeddingJSON), &mem.Embedding)

	return &mem, nil
}

// GetEmbedderDimensions returns the dimensions of the current embedder
func (s *Store) GetEmbedderDimensions() int {
	return s.embedder.Dimensions()
}

// Helper functions

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func hashString(s string) int {
	h := 0
	for _, c := range s {
		h = 31*h + int(c)
	}
	if h < 0 {
		h = -h
	}
	return h
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// ============================================================================
// Citation Methods
// ============================================================================

// AddCitation adds a citation to a memory
func (s *Store) AddCitation(ctx context.Context, memoryID string, filePath string, startLine, endLine int, commitSHA, content string) (*Citation, error) {
	id := generateID()
	now := time.Now()

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO citations (id, memory_id, file_path, start_line, end_line, commit_sha, content, confidence, verified_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 1.0, ?, ?)
	`, id, memoryID, filePath, startLine, endLine, commitSHA, content, now, now)

	if err != nil {
		return nil, fmt.Errorf("failed to add citation: %w", err)
	}

	return &Citation{
		ID:         id,
		MemoryID:   memoryID,
		FilePath:   filePath,
		StartLine:  startLine,
		EndLine:    endLine,
		CommitSHA:  commitSHA,
		Content:    content,
		Confidence: 1.0,
		VerifiedAt: now,
		CreatedAt:  now,
	}, nil
}

// GetCitations returns all citations for a memory
func (s *Store) GetCitations(ctx context.Context, memoryID string) ([]Citation, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, memory_id, file_path, start_line, end_line, commit_sha, content, confidence, verified_at, created_at
		FROM citations WHERE memory_id = ?
	`, memoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var citations []Citation
	for rows.Next() {
		var c Citation
		var commitSHA, content sql.NullString
		var verifiedAt sql.NullTime

		err := rows.Scan(&c.ID, &c.MemoryID, &c.FilePath, &c.StartLine, &c.EndLine, &commitSHA, &content, &c.Confidence, &verifiedAt, &c.CreatedAt)
		if err != nil {
			continue
		}

		if commitSHA.Valid {
			c.CommitSHA = commitSHA.String
		}
		if content.Valid {
			c.Content = content.String
		}
		if verifiedAt.Valid {
			c.VerifiedAt = verifiedAt.Time
		}

		citations = append(citations, c)
	}

	return citations, nil
}

// VerifyCitation checks if a citation is still valid by comparing file content
func (s *Store) VerifyCitation(ctx context.Context, citationID string) (*Citation, bool, error) {
	// Get the citation
	var c Citation
	var commitSHA, content sql.NullString
	var verifiedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, `
		SELECT id, memory_id, file_path, start_line, end_line, commit_sha, content, confidence, verified_at, created_at
		FROM citations WHERE id = ?
	`, citationID).Scan(&c.ID, &c.MemoryID, &c.FilePath, &c.StartLine, &c.EndLine, &commitSHA, &content, &c.Confidence, &verifiedAt, &c.CreatedAt)

	if err != nil {
		return nil, false, fmt.Errorf("citation not found: %w", err)
	}

	if commitSHA.Valid {
		c.CommitSHA = commitSHA.String
	}
	if content.Valid {
		c.Content = content.String
	}
	if verifiedAt.Valid {
		c.VerifiedAt = verifiedAt.Time
	}

	// Validate file path to prevent directory traversal
	if strings.Contains(c.FilePath, "..") {
		c.Confidence = 0.0
		s.updateCitationConfidence(ctx, c.ID, 0.0)
		return &c, false, fmt.Errorf("invalid file path: contains '..'")
	}

	// Check file size before reading (limit to 10MB)
	const maxFileSize = 10 * 1024 * 1024
	fileInfo, err := os.Stat(c.FilePath)
	if err != nil {
		c.Confidence = 0.0
		s.updateCitationConfidence(ctx, c.ID, 0.0)
		return &c, false, nil
	}
	if fileInfo.Size() > maxFileSize {
		c.Confidence = 0.0
		s.updateCitationConfidence(ctx, c.ID, 0.0)
		return &c, false, fmt.Errorf("file too large (%d bytes, limit %d)", fileInfo.Size(), maxFileSize)
	}

	// Read the current file content
	fileContent, err := os.ReadFile(c.FilePath)
	if err != nil {
		// File doesn't exist or can't be read - citation is invalid
		c.Confidence = 0.0
		s.updateCitationConfidence(ctx, c.ID, 0.0)
		return &c, false, nil
	}

	// If we have stored content, compare it
	if c.Content != "" {
		lines := strings.Split(string(fileContent), "\n")

		// Extract the cited lines
		startIdx := c.StartLine - 1
		endIdx := c.EndLine
		if startIdx < 0 {
			startIdx = 0
		}
		if endIdx > len(lines) {
			endIdx = len(lines)
		}
		if startIdx >= len(lines) {
			// Lines no longer exist
			c.Confidence = 0.0
			s.updateCitationConfidence(ctx, c.ID, 0.0)
			return &c, false, nil
		}

		currentContent := strings.Join(lines[startIdx:endIdx], "\n")

		// Check if content matches
		if strings.TrimSpace(currentContent) == strings.TrimSpace(c.Content) {
			// Content matches - citation is valid
			c.Confidence = 1.0
			c.VerifiedAt = time.Now()
			s.updateCitationVerified(ctx, c.ID, c.VerifiedAt, 1.0)
			return &c, true, nil
		}

		// Content changed - reduce confidence based on similarity
		similarity := stringSimilarity(c.Content, currentContent)
		c.Confidence = similarity
		s.updateCitationConfidence(ctx, c.ID, similarity)
		return &c, similarity > 0.8, nil
	}

	// No stored content - just check if file exists and lines are in range
	lines := strings.Split(string(fileContent), "\n")
	if c.StartLine > 0 && c.StartLine <= len(lines) {
		c.Confidence = 0.9 // File exists, lines in range, but can't verify content
		c.VerifiedAt = time.Now()
		s.updateCitationVerified(ctx, c.ID, c.VerifiedAt, 0.9)
		return &c, true, nil
	}

	c.Confidence = 0.0
	s.updateCitationConfidence(ctx, c.ID, 0.0)
	return &c, false, nil
}

// DecayCitations reduces confidence of all citations based on time since verification
// Decay rate: 10% per day since last verification
func (s *Store) DecayCitations(ctx context.Context) (int, error) {
	now := time.Now()

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, confidence, verified_at FROM citations WHERE confidence > 0
	`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	updated := 0
	for rows.Next() {
		var id string
		var confidence float64
		var verifiedAt sql.NullTime

		if err := rows.Scan(&id, &confidence, &verifiedAt); err != nil {
			continue
		}

		if !verifiedAt.Valid {
			continue
		}

		// Calculate days since verification
		daysSince := now.Sub(verifiedAt.Time).Hours() / 24
		if daysSince < 1 {
			continue
		}

		// Decay: 10% per day, minimum 0.1
		decayFactor := math.Pow(0.9, daysSince)
		newConfidence := confidence * decayFactor
		if newConfidence < 0.1 {
			newConfidence = 0.1
		}

		if newConfidence != confidence {
			s.updateCitationConfidence(ctx, id, newConfidence)
			updated++
		}
	}

	return updated, nil
}

// GetMemoryConfidence calculates aggregate confidence for a memory based on its citations
func (s *Store) GetMemoryConfidence(ctx context.Context, memoryID string) (float64, error) {
	var avgConfidence sql.NullFloat64
	err := s.db.QueryRowContext(ctx, `
		SELECT AVG(confidence) FROM citations WHERE memory_id = ?
	`, memoryID).Scan(&avgConfidence)

	if err != nil || !avgConfidence.Valid {
		return 1.0, nil // No citations = full confidence
	}

	return avgConfidence.Float64, nil
}

// Helper methods for updating citations
func (s *Store) updateCitationConfidence(ctx context.Context, id string, confidence float64) {
	s.db.ExecContext(ctx, `UPDATE citations SET confidence = ? WHERE id = ?`, confidence, id)
}

func (s *Store) updateCitationVerified(ctx context.Context, id string, verifiedAt time.Time, confidence float64) {
	s.db.ExecContext(ctx, `UPDATE citations SET verified_at = ?, confidence = ? WHERE id = ?`, verifiedAt, confidence, id)
}

// stringSimilarity calculates a simple similarity score between two strings
func stringSimilarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	// Simple word overlap similarity
	wordsA := strings.Fields(strings.ToLower(a))
	wordsB := strings.Fields(strings.ToLower(b))

	if len(wordsA) == 0 || len(wordsB) == 0 {
		return 0.0
	}

	// Count matching words
	wordSetB := make(map[string]bool)
	for _, w := range wordsB {
		wordSetB[w] = true
	}

	matches := 0
	for _, w := range wordsA {
		if wordSetB[w] {
			matches++
		}
	}

	// Jaccard-like similarity
	union := len(wordsA) + len(wordsB) - matches
	if union == 0 {
		return 0.0
	}

	return float64(matches) / float64(union)
}
