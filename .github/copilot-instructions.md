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

## Key Files

| Purpose | Go | C | TypeScript |
|---------|-----|---|-----------|
| **Crypto primitives** | [pkg/vax/vax.go](../go/pkg/vax/vax.go) | [c/src/gi.c](../c/src/gi.c), [sai.c](../c/src/sai.c) | — |
| **JCS canonicalization** | [internal/jcs/jcs.go](../go/internal/jcs/jcs.go) | — | [src/jcs.ts](../ts/src/jcs.ts) |
| **API header** | — | [include/vax.h](../c/include/vax.h) | — |
| **Spec** | [docs/SPECIFICATION.md](../docs/SPECIFICATION.md) | | |

## Workflow Commands

```bash
# Go — run all tests
cd go && go test ./...

# C — build and test
cd c && cmake -B build -G Ninja && cmake --build build && ctest --test-dir build

# TypeScript — test
cd ts && npm test
```

## Test Vectors (Cross-Language Verification)

```
# Genesis SAI — ALL implementations MUST produce this
actor_id: "user123:device456"
genesis_salt: a1a2a3a4a5a6a7a8a9aaabacadaeafb0
Expected SAI: afc50728cd79e805a8ae06875a1ddf78ca11b0d56ec300b160fb71f50ce658c3

# gi derivation (counter=1, zero k_chain)
Expected gi: 96b0dbcec77032023871b0df25214723e5b053da24d50b8f3338ea55f9966a69
```

## Common Pitfalls

| Mistake | Fix |
|---------|-----|
| Using `json.Marshal()` | Use `jcs.Marshal()` |
| Adding state to C core | State belongs in Go/TS layer |
| Adding JSON to C | C = crypto only |
| Wrong hash formula | Two-stage: `SHA256("VAX-SAI" \|\| prevSAI \|\| SHA256(SAE) \|\| gi)` |
| Little-endian counter | Counter MUST be big-endian uint16 |

## Design Principles (Why?)

1. **Tool, not protocol** — VAX doesn't prevent mistakes, it makes them impossible to quietly rewrite
2. **Single linear history per actor** — No cross-actor merging; divergence is detected, not resolved
3. **IRP (Inverse Responsibility)** — Producers normalize; backends only verify
4. **Pure Go preferred** — Go implementation has zero dependencies (no CGo)

## When in Doubt

| Question | Answer |
|----------|--------|
| "How should I encode this action?" | Use VAX-JCS (`jcs.Marshal()` in Go, `canonicalize()` in TS) |
| "How do I compute hashes?" | Go: `vax.ComputeGI()`, `vax.ComputeSAI()` / C: `vax_compute_*()` |
| "Should I add JSON handling to C?" | No. C = crypto only |
| "Should I merge histories?" | No. VAX detects divergence; resolution is L1+ |
| "Where's the consensus mechanism?" | There isn't one. See [ARCHITECTURE.md](../docs/ARCHITECTURE.md) §2 |
| "Can I modify SAE after creation?" | No. Actions are immutable once hashed |
