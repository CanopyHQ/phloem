package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRememberWithScope(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	// Store memories with different scopes
	mem1, err := store.RememberWithScope(ctx, "User authentication implemented", []string{"feature"}, "auth", "github.com/CanopyHQ/canopy")
	require.NoError(t, err)
	assert.Equal(t, "github.com/CanopyHQ/canopy", mem1.Scope)

	mem2, err := store.RememberWithScope(ctx, "Database schema updated", []string{"feature"}, "db", "github.com/CanopyHQ/phloem")
	require.NoError(t, err)
	assert.Equal(t, "github.com/CanopyHQ/phloem", mem2.Scope)

	mem3, err := store.RememberWithScope(ctx, "API endpoint added", []string{"feature"}, "api", "github.com/CanopyHQ/canopy")
	require.NoError(t, err)
	assert.Equal(t, "github.com/CanopyHQ/canopy", mem3.Scope)

	// Recall without scope filter - should get all memories
	results, err := store.Recall(ctx, "feature", 10, nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 3)

	// Recall with canopy scope - should only get canopy memories
	results, err = store.RecallWithScope(ctx, "feature", 10, nil, "github.com/CanopyHQ/canopy")
	require.NoError(t, err)
	assert.Equal(t, 2, len(results))
	for _, mem := range results {
		assert.Equal(t, "github.com/CanopyHQ/canopy", mem.Scope)
	}

	// Recall with phloem scope - should only get phloem memory
	results, err = store.RecallWithScope(ctx, "feature", 10, nil, "github.com/CanopyHQ/phloem")
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "github.com/CanopyHQ/phloem", results[0].Scope)
	assert.Contains(t, results[0].Content, "Database schema")
}

func TestRecallWithScopeAndTags(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	// Store memories with different scopes and tags
	_, err := store.RememberWithScope(ctx, "Bug fix in auth", []string{"bugfix"}, "auth", "github.com/CanopyHQ/canopy")
	require.NoError(t, err)

	_, err = store.RememberWithScope(ctx, "Feature in auth", []string{"feature"}, "auth", "github.com/CanopyHQ/canopy")
	require.NoError(t, err)

	_, err = store.RememberWithScope(ctx, "Bug fix in API", []string{"bugfix"}, "api", "github.com/CanopyHQ/phloem")
	require.NoError(t, err)

	// Recall with scope and tag filter
	results, err := store.RecallWithScope(ctx, "auth", 10, []string{"bugfix"}, "github.com/CanopyHQ/canopy")
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "github.com/CanopyHQ/canopy", results[0].Scope)
	assert.Contains(t, results[0].Content, "Bug fix in auth")

	// Recall with different scope
	results, err = store.RecallWithScope(ctx, "bug", 10, []string{"bugfix"}, "github.com/CanopyHQ/phloem")
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "github.com/CanopyHQ/phloem", results[0].Scope)
	assert.Contains(t, results[0].Content, "Bug fix in API")
}

func TestScopeIsolation(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	// Store identical content in different scopes
	content := "Implemented user authentication"

	mem1, err := store.RememberWithScope(ctx, content, []string{"auth"}, "security", "github.com/CanopyHQ/canopy")
	require.NoError(t, err)

	mem2, err := store.RememberWithScope(ctx, content, []string{"auth"}, "security", "github.com/CanopyHQ/phloem")
	require.NoError(t, err)

	// Should be different memories (different IDs)
	assert.NotEqual(t, mem1.ID, mem2.ID)

	// Recall from each scope separately
	canopyResults, err := store.RecallWithScope(ctx, "authentication", 10, nil, "github.com/CanopyHQ/canopy")
	require.NoError(t, err)
	assert.Equal(t, 1, len(canopyResults))
	assert.Equal(t, "github.com/CanopyHQ/canopy", canopyResults[0].Scope)

	phloemResults, err := store.RecallWithScope(ctx, "authentication", 10, nil, "github.com/CanopyHQ/phloem")
	require.NoError(t, err)
	assert.Equal(t, 1, len(phloemResults))
	assert.Equal(t, "github.com/CanopyHQ/phloem", phloemResults[0].Scope)
}

func TestEmptyScope(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()
	ctx := context.Background()

	// Store memory without scope (global)
	mem, err := store.Remember(ctx, "Global configuration", []string{"config"}, "global")
	require.NoError(t, err)
	assert.Empty(t, mem.Scope)

	// Store memory with scope
	scopedMem, err := store.RememberWithScope(ctx, "Project configuration", []string{"config"}, "project", "github.com/CanopyHQ/canopy")
	require.NoError(t, err)
	assert.Equal(t, "github.com/CanopyHQ/canopy", scopedMem.Scope)

	// Recall without scope filter - should get both
	results, err := store.Recall(ctx, "configuration", 10, nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2)

	// Recall with empty scope filter - same as no filter
	results, err = store.RecallWithScope(ctx, "configuration", 10, nil, "")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2)

	// Recall with specific scope - should only get scoped memory
	results, err = store.RecallWithScope(ctx, "configuration", 10, nil, "github.com/CanopyHQ/canopy")
	require.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "github.com/CanopyHQ/canopy", results[0].Scope)
}
