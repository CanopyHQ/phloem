package memory

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
)

func init() {
	sqlite_vec.Auto()
}

// vecIndex manages the sqlite-vec vector index for fast KNN queries.
// If the extension fails to load, all operations are no-ops and the store
// falls back to brute-force cosine similarity.
type vecIndex struct {
	db         *sql.DB
	dimensions int
	available  bool
}

type vecResult struct {
	MemoryID string
	Distance float64
}

func newVecIndex(db *sql.DB, dimensions int) *vecIndex {
	vi := &vecIndex{db: db, dimensions: dimensions}
	if err := vi.ensureSchema(); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  sqlite-vec not available, using linear scan: %v\n", err)
		vi.available = false
	} else {
		vi.available = true
	}
	return vi
}

func (vi *vecIndex) ensureSchema() error {
	// Verify vec0 extension is loaded
	var vecVersion string
	if err := vi.db.QueryRow("SELECT vec_version()").Scan(&vecVersion); err != nil {
		return fmt.Errorf("vec_version() failed: %w", err)
	}

	// Metadata table to track embedding dimensions
	if _, err := vi.db.Exec(`CREATE TABLE IF NOT EXISTS vec_metadata (key TEXT PRIMARY KEY, value TEXT)`); err != nil {
		return fmt.Errorf("failed to create vec_metadata: %w", err)
	}

	// ID mapping table (vec0 requires integer rowids, our memories use text IDs)
	if _, err := vi.db.Exec(`CREATE TABLE IF NOT EXISTS memory_vec_ids (
		vec_id INTEGER PRIMARY KEY AUTOINCREMENT,
		memory_id TEXT UNIQUE NOT NULL
	)`); err != nil {
		return fmt.Errorf("failed to create vec ID mapping: %w", err)
	}

	// Handle dimension changes (e.g. switching from local to OpenAI embedder)
	vi.handleDimensionChange()

	// Create vec0 virtual table with cosine distance
	createSQL := fmt.Sprintf(
		`CREATE VIRTUAL TABLE IF NOT EXISTS memory_embeddings USING vec0(embedding float[%d] distance_metric=cosine)`,
		vi.dimensions,
	)
	if _, err := vi.db.Exec(createSQL); err != nil {
		return fmt.Errorf("failed to create vec0 table: %w", err)
	}

	// Record current dimensions
	vi.db.Exec(`INSERT OR REPLACE INTO vec_metadata (key, value) VALUES ('dimensions', ?)`,
		fmt.Sprintf("%d", vi.dimensions))

	return nil
}

// handleDimensionChange detects if the embedder dimensions changed since last run
// and drops the vec0 table so it can be recreated with the correct dimensions.
func (vi *vecIndex) handleDimensionChange() {
	var storedDim string
	err := vi.db.QueryRow(`SELECT value FROM vec_metadata WHERE key = 'dimensions'`).Scan(&storedDim)
	if err != nil {
		return // No stored dimensions yet, first run
	}
	if storedDim == fmt.Sprintf("%d", vi.dimensions) {
		return // Dimensions match
	}

	// Dimension mismatch - drop and recreate
	fmt.Fprintf(os.Stderr, "⚠️  Embedding dimensions changed (%s -> %d), rebuilding vec index\n", storedDim, vi.dimensions)
	vi.db.Exec(`DROP TABLE IF EXISTS memory_embeddings`)
	vi.db.Exec(`DELETE FROM memory_vec_ids`)
}

// Insert adds or replaces a memory's embedding in the vec0 index.
func (vi *vecIndex) Insert(memoryID string, embedding []float32) error {
	if !vi.available || len(embedding) == 0 || len(embedding) != vi.dimensions {
		return nil
	}

	// Get or create vec_id for this memory
	var vecID int64
	err := vi.db.QueryRow(`SELECT vec_id FROM memory_vec_ids WHERE memory_id = ?`, memoryID).Scan(&vecID)
	if err == sql.ErrNoRows {
		result, err := vi.db.Exec(`INSERT INTO memory_vec_ids (memory_id) VALUES (?)`, memoryID)
		if err != nil {
			return fmt.Errorf("failed to create vec ID mapping: %w", err)
		}
		vecID, _ = result.LastInsertId()
	} else if err != nil {
		return err
	}

	blob, err := sqlite_vec.SerializeFloat32(embedding)
	if err != nil {
		return fmt.Errorf("failed to serialize embedding: %w", err)
	}

	// vec0 doesn't support ON CONFLICT, so delete first if exists
	vi.db.Exec(`DELETE FROM memory_embeddings WHERE rowid = ?`, vecID)

	_, err = vi.db.Exec(`INSERT INTO memory_embeddings (rowid, embedding) VALUES (?, ?)`, vecID, blob)
	if err != nil {
		return fmt.Errorf("failed to insert into vec0: %w", err)
	}

	return nil
}

// Search performs a KNN query and returns memory IDs with cosine distances.
func (vi *vecIndex) Search(queryEmbedding []float32, limit int) ([]vecResult, error) {
	if !vi.available {
		return nil, fmt.Errorf("vec index not available")
	}

	blob, err := sqlite_vec.SerializeFloat32(queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize query: %w", err)
	}

	// Step 1: KNN query on vec0 (returns rowids + distances)
	rows, err := vi.db.Query(`
		SELECT rowid, distance
		FROM memory_embeddings
		WHERE embedding MATCH ?
		ORDER BY distance
		LIMIT ?
	`, blob, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type rowResult struct {
		rowID    int64
		distance float64
	}
	var rowResults []rowResult
	for rows.Next() {
		var r rowResult
		if err := rows.Scan(&r.rowID, &r.distance); err != nil {
			continue
		}
		rowResults = append(rowResults, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(rowResults) == 0 {
		return nil, nil
	}

	// Step 2: Batch-map rowids to memory_ids
	placeholders := make([]string, len(rowResults))
	args := make([]interface{}, len(rowResults))
	for i, rr := range rowResults {
		placeholders[i] = "?"
		args[i] = rr.rowID
	}

	mapRows, err := vi.db.Query(
		`SELECT vec_id, memory_id FROM memory_vec_ids WHERE vec_id IN (`+strings.Join(placeholders, ",")+`)`,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer mapRows.Close()

	idMap := make(map[int64]string)
	for mapRows.Next() {
		var vecID int64
		var memID string
		if err := mapRows.Scan(&vecID, &memID); err != nil {
			continue
		}
		idMap[vecID] = memID
	}

	// Build results preserving KNN order
	var results []vecResult
	for _, rr := range rowResults {
		if memID, ok := idMap[rr.rowID]; ok {
			results = append(results, vecResult{MemoryID: memID, Distance: rr.distance})
		}
	}

	return results, nil
}

// Delete removes a memory from the vec0 index.
func (vi *vecIndex) Delete(memoryID string) error {
	if !vi.available {
		return nil
	}
	var vecID int64
	if err := vi.db.QueryRow(`SELECT vec_id FROM memory_vec_ids WHERE memory_id = ?`, memoryID).Scan(&vecID); err != nil {
		return nil // Not in vec index
	}
	vi.db.Exec(`DELETE FROM memory_embeddings WHERE rowid = ?`, vecID)
	vi.db.Exec(`DELETE FROM memory_vec_ids WHERE vec_id = ?`, vecID)
	return nil
}

// Backfill populates the vec0 index from existing memories that have embeddings.
// Returns the number of memories backfilled.
func (vi *vecIndex) Backfill(db *sql.DB) (int, error) {
	if !vi.available {
		return 0, nil
	}

	// Check if backfill is needed
	var vecCount int
	vi.db.QueryRow(`SELECT COUNT(*) FROM memory_vec_ids`).Scan(&vecCount)

	var memCount int
	db.QueryRow(`SELECT COUNT(*) FROM memories WHERE embedding IS NOT NULL AND embedding != '' AND embedding != '[]' AND embedding != 'null'`).Scan(&memCount)

	if vecCount >= memCount || memCount == 0 {
		return 0, nil
	}

	// Fetch memories with embeddings that are not yet in the vec index
	rows, err := db.Query(`
		SELECT m.id, m.embedding
		FROM memories m
		LEFT JOIN memory_vec_ids v ON v.memory_id = m.id
		WHERE v.vec_id IS NULL
		AND m.embedding IS NOT NULL AND m.embedding != '' AND m.embedding != '[]' AND m.embedding != 'null'
	`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var memID, embJSON string
		if err := rows.Scan(&memID, &embJSON); err != nil {
			continue
		}

		var embedding []float32
		if err := json.Unmarshal([]byte(embJSON), &embedding); err != nil {
			continue
		}

		if len(embedding) != vi.dimensions {
			continue // Skip mismatched dimensions
		}

		if err := vi.Insert(memID, embedding); err != nil {
			continue
		}
		count++
	}

	return count, nil
}
