# VAX Action History — AI Coding Instructions

## Project Overview

VAX is a **Git-like tool for tamper-evident action histories**. It provides deterministic, actor-bound action logs with cryptographic verification. Local-first, no blockchain required.

**Core metaphor:** Git is to code what VAX is to actions.

## Architecture at a Glance

```
┌─────────────────────────────────────────────────────────────┐
│  Language Implementations (all produce identical outputs)   │
├─────────────┬─────────────────┬─────────────────────────────┤
│  C (crypto) │  Go (pure)      │  TypeScript                 │
│  libvax.a   │  pkg/vax        │  ts/src/jcs.ts              │
│  OpenSSL    │  zero deps      │  zero deps                  │
└─────────────┴─────────────────┴─────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│  Hash Chain: SAI_n = SHA256("VAX-SAI" || prevSAI ||         │
│                              SHA256(SAE) || gi_n)           │
└─────────────────────────────────────────────────────────────┘
```

## Critical Rules

### 1. NEVER Use Standard JSON Serialization
```go
// ❌ WRONG
jsonBytes, _ := json.Marshal(action)

// ✅ CORRECT — Always use VAX-JCS
import "vax/internal/jcs"
saeBytes, err := jcs.Marshal(action)
```

### 2. C Core = Pure Crypto ONLY
- ✅ `vax_compute_gi()`, `vax_compute_sai()`, `vax_verify_action()`
- ❌ No JSON handling, no state management, no `vax_chain_*` functions

### 3. Go/TS Handle JSON + State
- JSON canonicalization → `go/internal/jcs/` or `ts/src/jcs.ts`
- State tracking (counter, prevSAI) → application layer
# VAX — AI Coding Notes (concise)

Purpose
- VAX is a multi-language, deterministic action-history tool: C implements crypto primitives; Go/TS implement canonicalization, state and application logic.

Key rules (must-follow)
- C: crypto only — no JSON, no state, no chain logic (`c/src/gi.c`, `c/src/sai.c`).
- Canonicalize actions with VAX-JCS: `go/internal/jcs/jcs.go` or `ts/src/jcs.ts` (do NOT use `json.Marshal`).
- Hash formulas: `SAI_n = SHA256("VAX-SAI" || prevSAI || SHA256(SAE) || gi_n)`; `gi = HMAC_SHA256(k_chain, "VAX-GI" || counter)` with counter as big-endian uint16.

Quick workflows (copy-paste)
- Go tests: `cd go && go test ./...`
- C build+tests: `cd c && cmake -B build -G Ninja && cmake --build build && ctest --test-dir build`
- TS tests: `cd ts && npm test`

Where to edit safely
- Add JSON/state or JCS changes in Go/TS only. See `go/pkg/vax/vax.go` for hash helpers.
- If changing C, limit edits to pure crypto functions and run `c/test` unit tests.

Patterns & examples
- Use `jcs.Marshal()` (Go) or `canonicalize()` (TS) before hashing SAE.
- Tests and cross-language vectors live under `go/internal/jcs/`, `ts/src/jcs.*`, and `test-vectors.json`.

Common pitfalls
- Accidentally using native JSON → nondeterministic outputs across languages.
- Encoding the counter as little-endian (must be big-endian uint16).

Helpful references
- Spec & algorithms: `docs/SPECIFICATION.md`
- JCS canonicalization: `go/internal/jcs/jcs.go`, `ts/src/jcs.ts`
- Crypto primitives: `go/pkg/vax/vax.go`, `c/src/gi.c`, `c/src/sai.c`

If anything in these notes is unclear or you want more examples (small patches showing correct `jcs.Marshal()` usage or GI/SAI tests), tell me which area to expand.
