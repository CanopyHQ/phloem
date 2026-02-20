# Phloem Grafts (DShare): The Viral Engine

**Status**: Design
**Priority**: P0 (Viral Requirement)

## 🎯 The "Viral Pull" Thesis
If Canopy is just a local memory tool, it grows linearly.
If Canopy is the *standard format* for sharing "Context Bundles" (Grafts), it grows exponentially.

**The Loop**:
1. Expert Developer curates memories about a topic (e.g., "Stripe Integration patterns").
2. Expert exports a **Graft**: `stripe-patterns.graft`.
3. Expert shares Graft on Twitter/Team Slack.
4. Junior Developer sees it: "I need that context."
5. Junior installs Canopy (`brew install canopy`) to consume the Graft.
6. Junior now has Canopy and starts generating their own context.

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
    "name": "Canopy Architecture Patterns",
    "description": "Core patterns for Crown, Phloem, and Cambium development.",
    "author": "Duncan Rose",
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
phloem graft export --tags "architecture,patterns" --output canopy-arch.graft

# Export recent session
phloem graft export --since "24h" --output daily-sync.graft
```

#### Import
```bash
# Import a graft file
phloem graft import canopy-arch.graft

# Output:
# 📦 Reading "Canopy Architecture Patterns"...
# 🔓 Verifying format... OK
# ✨ Imported 45 memories by Duncan Rose
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
- **Seed Content**: "Canopy Engineering Standards" graft.

### Phase 2: The Registry (Post-Beta)
- `phloem graft publish` -> Uploads to Crown.
- `phloem graft install canopy/architecture` -> Downloads from Crown.

---

## ⚠️ Implementation Details
- **Package**: `github.com/CanopyHQ/phloem/internal/graft`
- **Magic Bytes**: `[]byte{0x50, 0x48, 0x4C, 0x4F}`
- **Compression**: `compress/gzip`
