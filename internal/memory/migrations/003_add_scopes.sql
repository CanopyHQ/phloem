-- Add scope support to memories table
-- Scope represents the context boundary (e.g., GitHub repo) for a memory
-- NULL scope = global/personal memory
-- Non-NULL scope = scoped to specific repo/project

ALTER TABLE memories ADD COLUMN scope TEXT;

-- Create index for scope-based queries
CREATE INDEX IF NOT EXISTS idx_memories_scope ON memories(scope);

-- Create scopes table to track available scopes
CREATE TABLE IF NOT EXISTS scopes (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,           -- Human-readable name (e.g., "CanopyHQ/canopy")
    type TEXT NOT NULL,            -- "github_repo", "project", "workspace"
    metadata TEXT,                 -- JSON metadata (repo URL, owner, etc.)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index for scope lookups
CREATE INDEX IF NOT EXISTS idx_scopes_name ON scopes(name);
CREATE INDEX IF NOT EXISTS idx_scopes_type ON scopes(type);
