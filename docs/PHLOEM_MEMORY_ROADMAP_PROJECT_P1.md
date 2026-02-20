# Phloem Memory Roadmap Implementation (P1)

**Priority:** P1  
**Branch:** `feat/phloem-memory-stages`  
**Status:** Ready for pickup

## Scope

Implement the full memory research roadmap (Stage 1 → Stage 2 → Stage 3): edge table and causal relation extraction (Stage 1), causal DAG queries and compositional compose plus prefetch/session preload and memory dreams (Stage 2), memory critic/reorg and air-gapped mode and local embeddings (Stage 3). Full step-by-step plan: [docs/PHLOEM_MEMORY_ROADMAP_IMPLEMENTATION_PLAN.md](../../docs/PHLOEM_MEMORY_ROADMAP_IMPLEMENTATION_PLAN.md).

## Agent Assignments

| Agent | Share | Ownership | Deliverables |
|-------|-------|-----------|---------------|
| **Windsurf (high-level)** | ~28% | Phase 0 + Stage 3 + coordination | Branch strategy; roadmap doc updates; Stage 3 (memory critic/reorg, air-gapped, local embeddings); integration review and definition-of-done per stage |
| **Cursor 1 (mid-level)** | ~36% | Stage 1 | Edge table schema; causal relation extraction pipeline; wire to ingest; temporal edges; tests and docs for Stage 1 |
| **Cursor 2 (mid-level)** | ~36% | Stage 2 | DAG causal queries + MCP; compositional compose + MCP; predictive prefetch + session preload; memory dreams; tests and docs for Stage 2 |

## Dependencies

- Phase 0 (Windsurf) first: create branch `feat/phloem-memory-stages`, update [docs/Roadmap-as-of-Jan21-PM.md](../../docs/Roadmap-as-of-Jan21-PM.md) Phase 1 Parity to Done.
- Cursor 1 delivers Stage 1 before Cursor 2 implements DAG/compose.
- Cursor 2 delivers Stage 2 before Windsurf implements Stage 3 (local embeddings can integrate `opus-s/feat/on-device-embeddings`).

## Definition of done

- **Stage 1:** `memory_edges` table exists; causal extraction pipeline runs after Remember/ingest; temporal edges on insert; unit tests and docs updated.
- **Stage 2:** Causal queries and compose available via MCP; session preload and prefetch improve context; optional memory-dreams pass runs offline.
- **Stage 3:** Memory critic and reorganization run; air-gapped mode and local embeddings available; roadmap and DEEP_RESEARCH Stage 3 marked implemented.

## Test gates (before merging each stage)

- Run Phloem tests: `cd phloem && go test ./internal/... ./cmd/... -count=1`
- Optional: run `opus-s/feat/zero-defect-tests` and `opus-s/qa/crown-memories-tests` branches' checks before release

## References

- Full plan: [docs/PHLOEM_MEMORY_ROADMAP_IMPLEMENTATION_PLAN.md](../../docs/PHLOEM_MEMORY_ROADMAP_IMPLEMENTATION_PLAN.md)
- Research stages: [docs/DEEP_RESEARCH_MEMORY_SYSTEMS.md](../../docs/DEEP_RESEARCH_MEMORY_SYSTEMS.md)
- Phloem v2 architecture: [phloem/docs/PHLOEM_V2_ARCHITECTURE.md](PHLOEM_V2_ARCHITECTURE.md)
