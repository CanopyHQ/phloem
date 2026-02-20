# Phloem Grafts (DShare): The Viral Engine

**Status**: Design
**Priority**: P0 (Viral Requirement)

## 🎯 The "Viral Pull" Thesis
If Phloem is just a local memory tool, it grows linearly.
If Phloem is the *standard format* for sharing "Context Bundles" (Grafts), it grows exponentially.

**The Loop**:
1. Expert Developer curates memories about a topic (e.g., "Stripe Integration patterns").
2. Expert exports a **Graft**: `stripe-patterns.graft`.
3. Expert shares Graft on Twitter/Team Slack.
4. Junior Developer sees it: "I need that context."
5. Junior installs Phloem (`brew install canopyhq/tap/phloem`) to consume the Graft.
6. Junior now has Phloem and starts generating their own context.

---

## 🏗 Technical Design: The Opaque Container

To prevent "bike-shedding" and emphasize that Grafts are software artifacts, we use an **Opaque Container** format.

### 1. File Format (`.graft`)
A custom binary format consisting of:
1.  **Magic Bytes** (4 bytes): `0x50 0x48 0x4C 0x4F` ("PHLO")
2.  **Version** (1 byte): `0x01`
3.  **Compression**: Gzip-compressed JSON payload.

**Visual Representation**:
```
[PHLO] [01] [ GZIP COMPRESSED PAYLOAD (Manifest + Memories) ]
```

### 2. Payload Schema (JSON)
Inside the compressed block is a single JSON object:

```json
{
  "manifest": {
    "id": "graft_12345",
    "name": "Phloem Architecture Patterns",
    "description": "Core patterns for Phloem development.",
    "author": "Phloem Team",
    "version": "1.0.0",
    "created_at": "2026-01-26T12:00:00Z",
    "memory_count": 45,
    "tags": ["architecture", "go", "patterns"]
  },
  "memories": [
    {
      "content": "...",
      "tags": ["..."],
      "created_at": "..."
    }
  ],
  "citations": []
}
```

### 3. CLI Commands

#### Export
```bash
# Export all memories with specific tags
phloem graft export --tags "architecture,patterns" --output phloem-arch.graft

# Export recent session
phloem graft export --since "24h" --output daily-sync.graft
```

#### Import
```bash
# Import a graft file
phloem graft import phloem-arch.graft

# Output:
# 📦 Reading "Phloem Architecture Patterns"...
# 🔓 Verifying format... OK
# ✨ Imported 45 memories by Phloem Team
```

### 4. Safety & Trust
- **Sandboxing**: Grafts are data-only. No executable code.
- **Review**: `phloem graft inspect <file>` shows manifest without importing.
- **Deduplication**: Import checks content hashes to avoid duplicates.

---

## 🚀 Implementation Plan

### Phase 1: The "Sneakernet" (Beta Launch)
- Binary `.graft` format implementation.
- `phloem graft export/import` commands.
- **Seed Content**: "Phloem Engineering Standards" graft.

### Phase 2: The Registry (Planned - Not yet implemented)
- `phloem graft publish` -> Uploads to registry. *(Not yet implemented)*
- `phloem graft install phloem/architecture` -> Downloads from registry. *(Not yet implemented)*

---

## ⚠️ Implementation Details
- **Package**: `github.com/CanopyHQ/phloem/internal/graft`
- **Magic Bytes**: `[]byte{0x50, 0x48, 0x4C, 0x4F}`
- **Compression**: `compress/gzip`
