# Phloem Grafts Tutorial

**Grafts** are shareable memory bundles - the viral engine of Canopy. Export your curated context and share it with others.

## What Are Grafts?

A **graft** is a `.graft` file containing:
- Curated memories on a specific topic
- Metadata (name, description, author)
- Optional citations

Think of it as a "knowledge package" you can share.

## Quick Start

### Export a Graft

```bash
# Export all memories tagged with "architecture"
canopy graft export --tags "architecture,patterns" --output my-arch.graft --name "My Architecture Patterns"

# Export recent memories (last 24 hours)
canopy graft export --since "24h" --output daily-sync.graft
```

### Share a Graft

1. **Upload to GitHub**: Add to a release or repository
2. **Share on Twitter/X**: "Check out my Canopy graft: [link]"
3. **Share in Slack**: Post the file in your team channel

### Import a Graft

```bash
# First, inspect without importing
canopy graft inspect someone-else.graft

# If it looks good, import it
canopy graft import someone-else.graft
```

## Use Cases

### 1. Team Onboarding

**Export**: Your team's engineering standards
```bash
canopy graft export --tags "engineering,standards,onboarding" \
  --output team-standards.graft \
  --name "Team Engineering Standards" \
  --desc "Our coding standards and best practices"
```

**Share**: New team members import to get instant context

### 2. Project Knowledge Transfer

**Export**: All memories about a specific project
```bash
canopy graft export --tags "project-x,decisions,architecture" \
  --output project-x-context.graft
```

**Share**: Hand off to the next developer

### 3. Expert Knowledge Sharing

**Export**: Your expertise on a topic
```bash
canopy graft export --tags "stripe,payments,integration" \
  --output stripe-patterns.graft \
  --name "Stripe Integration Patterns" \
  --author "Your Name"
```

**Share**: Help others learn from your experience

## Best Practices

### 1. Curate Before Exporting

Don't export everything - be selective:
- Focus on a specific topic
- Use tags to filter
- Review memories before exporting

### 2. Write Good Descriptions

Help others understand what's in the graft:
```bash
--desc "Core patterns for building scalable Go services. Includes error handling, testing, and deployment strategies."
```

### 3. Version Your Grafts

Update grafts as you learn:
```bash
canopy graft export --tags "patterns" --output patterns-v2.graft --name "Patterns v2.0"
```

### 4. Include Citations

Grafts preserve citations to source files:
- Export memories that have citations
- Citations help verify and update knowledge

## The Viral Loop

1. **Expert** exports a graft (`stripe-patterns.graft`)
2. **Expert** shares on Twitter/Slack
3. **Learner** sees it: "I need that!"
4. **Learner** installs Canopy (`brew install canopy`) to import it
5. **Learner** now has Canopy and starts creating their own context
6. **Learner** exports their own grafts
7. **Loop continues** - exponential growth

## Example: Creating a Seed Graft

```bash
# 1. Tag your best memories
# (Memories should already be tagged with "architecture", "patterns", etc.)

# 2. Export
canopy graft export \
  --tags "architecture,patterns,engineering" \
  --output canopy-engineering-standards.graft \
  --name "Canopy Engineering Standards" \
  --desc "Core patterns for Crown, Phloem, and Cambium development" \
  --author "Canopy Team"

# 3. Verify
canopy graft inspect canopy-engineering-standards.graft

# 4. Share
# Upload to GitHub, share link, etc.
```

## Security & Trust

### Inspect Before Importing

Always inspect grafts before importing:
```bash
canopy graft inspect unknown.graft
```

This shows:
- Manifest (name, description, author)
- Memory count
- Tags
- **No memories are imported** - safe to inspect

### Grafts Are Data-Only

- Grafts contain **only memories** (text data)
- **No executable code**
- **No system access**
- Safe to import from trusted sources

### Deduplication

Canopy automatically deduplicates on import:
- If you already have a memory with the same content, it won't be duplicated
- Only unique memories are added

## Advanced Usage

### Export with Citations

Citations are automatically included if memories have them:
```bash
# Memories with citations will include citation data in the graft
canopy graft export --tags "documented" --output with-citations.graft
```

### Filter by Time

Export only recent memories:
```bash
# Last 7 days
canopy graft export --since "7d" --output recent.graft

# Last 2 weeks
canopy graft export --since "336h" --output biweekly.graft
```

### Combine Filters

```bash
# Recent architecture decisions
canopy graft export \
  --tags "architecture,decision" \
  --since "30d" \
  --output recent-arch-decisions.graft
```

## Troubleshooting

### "No memories matched criteria"

**Problem**: Export found no memories

**Solution**:
1. Check your tags: `canopy status` (shows memory counts)
2. Verify memories have the tags you're filtering by
3. Try without tags to see all memories: `canopy graft export --output all.graft`

### Import Fails

**Problem**: Graft file won't import

**Solution**:
1. Inspect first: `canopy graft inspect file.graft`
2. Check file format: Should start with magic bytes `PHLO`
3. Verify version compatibility

### Graft File Too Large

**Problem**: Graft file is very large

**Solution**:
- Use more specific tags to filter
- Export by time range (`--since`)
- Split into multiple grafts by topic

## Next Steps

1. **Try the seed graft**: `canopy graft import grafts/canopy-engineering-standards.graft`
2. **Create your first graft**: Export your expertise
3. **Share it**: Help others learn
4. **Build the viral loop**: Every graft shared = more Canopy users

---

**Remember**: Grafts are the viral engine. Every graft you share helps Canopy grow. Share your knowledge!
