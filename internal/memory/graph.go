// Package memory: DAG layer for Stage 2 (compose/prefetch). Cursor 1 populates causal edges; Cursor 2 uses for composition.
// Edge type and AddEdge/GetEdgesFrom/GetEdgesTo live in store.go; this file provides convenience wrappers.

package memory

import "context"

// EdgesFrom returns edges originating from the given memory ID (delegates to GetEdgesFrom with no type filter).
func (s *Store) EdgesFrom(ctx context.Context, memoryID string) ([]Edge, error) {
	return s.GetEdgesFrom(ctx, memoryID, "")
}

// EdgesTo returns edges pointing to the given memory ID (delegates to GetEdgesTo with no type filter).
func (s *Store) EdgesTo(ctx context.Context, memoryID string) ([]Edge, error) {
	return s.GetEdgesTo(ctx, memoryID, "")
}
