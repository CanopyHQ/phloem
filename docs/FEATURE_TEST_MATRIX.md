# Phloem Feature-Test Matrix

> **Last Updated**: 2026-01-12
> **Test Coverage Target**: 80%+ unit, 60%+ integration

This document maps every documented feature to its corresponding tests, ensuring zero-defect releases.

## Legend

| Symbol | Meaning |
|--------|---------|
| ✅ | Fully tested |
| ⚠️ | Partially tested |
| ❌ | Not tested |
| 🔄 | In progress |

---

## Memory Store (`internal/memory/store.go`)

| Feature | Unit Test | Integration | Edge Cases | Status |
|---------|-----------|-------------|------------|--------|
| **Create memory** | `TestRemember_Basic` | - | `TestRemember_EmptyContent`, `TestRemember_SpecialCharacters`, `TestRemember_LongContent` | ✅ |
| **Unique IDs** | `TestRemember_UniqueIDs` | - | - | ✅ |
| **Tag support** | `TestRemember_MultipleTags` | - | `TestRemember_VeryLongTags`, `TestRemember_ManyTags` | ✅ |
| **Context storage** | `TestRemember_Basic` | - | - | ✅ |
| **Semantic search** | `TestRecall_Basic` | - | `TestRecall_EmptyStore`, `TestRecall_SortedBySimilarity` | ✅ |
| **Tag filtering** | `TestRecall_WithTagFilter`, `TestList_WithTagFilter` | - | - | ✅ |
| **Result limiting** | `TestRecall_LimitResults`, `TestList_WithLimit` | - | - | ✅ |
| **Delete memory** | `TestForget_Basic` | - | `TestForget_NonExistent`, `TestForget_RemovesTags` | ✅ |
| **List memories** | `TestList_Basic` | - | - | ✅ |
| **Count memories** | `TestCount_Empty`, `TestCount_AfterOperations` | - | - | ✅ |
| **Database size** | `TestSize_ReturnsReadableString` | - | - | ✅ |
| **Last activity** | `TestLastActivity_Empty`, `TestLastActivity_AfterRemember` | - | - | ✅ |
| **Concurrent access** | `TestConcurrentRemember`, `TestConcurrentRecall` | - | - | ✅ |

### Embedding System

| Feature | Unit Test | Notes |
|---------|-----------|-------|
| **TF-IDF generation** | `TestGenerateEmbedding_Deterministic`, `TestGenerateEmbedding_DifferentTexts`, `TestGenerateEmbedding_EmptyText` | Local-only, no API calls |
| **Cosine similarity** | `TestCosineSimilarity_Identical`, `TestCosineSimilarity_Orthogonal`, `TestCosineSimilarity_DifferentLengths`, `TestCosineSimilarity_Empty` | Pure math functions |

---

## MCP Server (`internal/mcp/server.go`)

| Feature | Unit Test | Integration | Edge Cases | Status |
|---------|-----------|-------------|------------|--------|
| **Server creation** | `TestNewServer` | - | - | ✅ |
| **MCP initialize** | `TestHandleInitialize` | - | Protocol version, capabilities | ✅ |
| **Tools list** | `TestHandleToolsList`, `TestToolsHaveValidSchema` | - | All 5 tools registered | ✅ |
| **Tool: remember** | `TestToolCall_Remember` | - | `TestToolCall_Remember_MissingContent` | ✅ |
| **Tool: recall** | `TestToolCall_Recall` | - | `TestToolCall_Recall_MissingQuery`, `TestToolCall_Recall_WithTagFilter` | ✅ |
| **Tool: forget** | `TestToolCall_Forget` | - | `TestToolCall_Forget_MissingID` | ✅ |
| **Tool: list_memories** | `TestToolCall_ListMemories` | - | - | ✅ |
| **Tool: memory_stats** | `TestToolCall_MemoryStats` | - | - | ✅ |
| **Unknown tool handling** | `TestToolCall_UnknownTool` | - | Error code -32602 | ✅ |
| **Resources list** | `TestHandleResourcesList` | - | - | ✅ |
| **Resource: recent** | `TestHandleResourceRead_RecentMemories` | - | - | ✅ |
| **Resource: stats** | `TestHandleResourceRead_Stats` | - | - | ✅ |
| **Unknown resource** | `TestHandleResourceRead_UnknownURI` | - | - | ✅ |
| **Prompts list** | `TestHandlePromptsList` | - | - | ✅ |
| **Unknown method** | `TestUnknownMethod` | - | Error code -32601 | ✅ |
| **Invalid params** | `TestInvalidParams` | - | - | ✅ |
| **Stats helper** | `TestGetMemoryStats`, `TestGetMemoryStats_Empty` | - | - | ✅ |
| **Truncate helper** | `TestTruncate` | - | Various lengths | ✅ |

---

## Cloud Sync (`internal/sync/client.go`)

| Feature | Unit Test | Integration | Edge Cases | Status |
|---------|-----------|-------------|------------|--------|
| **Config loading** | - | - | - | ⚠️ Needs tests |
| **Sync upload** | - | - | - | ⚠️ Needs tests |
| **Sync download** | - | - | - | ⚠️ Needs tests |
| **API key auth** | - | - | - | ⚠️ Needs tests |
| **Error handling** | - | - | - | ⚠️ Needs tests |

---

## CLI Commands (`cmd/root.go`)

| Feature | Unit Test | Integration | Edge Cases | Status |
|---------|-----------|-------------|------------|--------|
| **serve** | - | Manual MCP test | - | ⚠️ Needs tests |
| **sync** | - | - | - | ⚠️ Needs tests |
| **list** | - | - | - | ⚠️ Needs tests |
| **forget** | - | - | - | ⚠️ Needs tests |
| **stats** | - | - | - | ⚠️ Needs tests |
| **import-chatgpt** | - | - | - | ⚠️ Needs tests |
| **import-claude** | - | - | - | ⚠️ Needs tests |

---

## Test Coverage Summary

| Package | Tests | Coverage | Target |
|---------|-------|----------|--------|
| `internal/memory` | 36 | ~90% | 80% ✅ |
| `internal/mcp` | 25 | ~85% | 80% ✅ |
| `internal/sync` | 0 | 0% | 60% ❌ |
| `cmd` | 0 | 0% | 60% ❌ |

---

## Running Tests

```bash
# All tests
go test ./... -v

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Specific package
go test ./internal/memory/... -v

# Race detection
go test ./... -race
```

---

## Adding New Tests

When adding a new feature:

1. Add entry to this matrix FIRST
2. Write tests BEFORE implementing
3. Update status when tests pass
4. Run full test suite before committing

## Test Requirements for Release

Before any release:

- [ ] All unit tests pass
- [ ] Coverage >= 80% for core packages
- [ ] No race conditions (`go test -race`)
- [ ] Edge cases documented and tested
- [ ] This matrix is up to date
