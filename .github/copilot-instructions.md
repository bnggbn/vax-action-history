# VAX Action History — AI Coding Instructions

## Project Overview

VAX is a **Git-like tool for tamper-evident action histories**, not a distributed consensus protocol. It provides deterministic, actor-bound action logs with cryptographic verification. Local-first, incrementally adoptable, no blockchain required.

**Core metaphor:** Git is to code what VAX is to actions.

## Key Design Principles

### 1. Tool, Not Protocol
VAX is a reference implementation, not a normative standard. It doesn't prevent mistakes—it makes them impossible to quietly rewrite. See [docs/ARCHITECTURE.md](../docs/ARCHITECTURE.md) for the "why" behind this.

### 2. Single Linear History Per Actor
Each `(user_id, device_id)` tuple = one strictly linear action chain. No cross-actor merging, no distributed consensus. Divergence is detected, not resolved—that's L0's job.

### 3. IRP — Inverse Responsibility Principle
Semantics are backend-defined, but normalization happens **producer-side**. Backends verify correctness; they never repair or mutate. This ensures cross-language/cross-version reproducibility.

## Critical Implementation Rules

### Canonical JSON (VAX-JCS)
**NEVER use `json.Marshal()` or standard JSON encoding for action data.**

- Use [go/internal/jcs/jcs.go](../go/internal/jcs/jcs.go) (Go) or [c/include/vax.h](../c/include/vax.h) `vax_canonicalize_json()` (C)
- **Requirements:** UTF-8, no whitespace, lexicographic key order, no scientific notation, non-ASCII escaped as `\uXXXX`
- See [docs/SPECIFICATION.md](../docs/SPECIFICATION.md) §2 for exact format rules

**Example pattern in Go:**
```go
// ❌ WRONG — Don't do this
jsonBytes, _ := json.Marshal(action)

// ✅ CORRECT — Always use JCS
import "vax/internal/jcs"
saeBytes, err := jcs.Marshal(action)
```

### Hash Chain Construction
Actions form an append-only chain where each SAI (Semantic Action Identifier) depends on:
```
SAI_n = SHA256("VAX-SAI" || prevSAI || SAE || gi_n)
```
- `gi_n`: per-action entropy via `HMAC_SHA256(k_chain, "VAX-GI" || counter)`
- `prevSAI`: previous action's hash (32 bytes)
- `SAE`: canonical JSON from VAX-JCS

**Never compute hashes manually**—use:
- C: `vax_compute_sai()` in [c/include/vax.h](../c/include/vax.h)
- Go: Implement following the C reference (not yet in Go codebase)

### State Management
The C API has two layers:
1. **Pure functions** (no state): `vax_compute_gi()`, `vax_compute_sai()`, `vax_canonicalize_json()`
2. **Stateful chain API**: `vax_chain_t*` handles counter/prevSAI tracking

When implementing new features, prefer pure functions for testability. Use `vax_chain_append()` for high-level operations.

## Testing Strategy

### Go Tests
- Run: `cd go && go test ./...`
- Key test: [go/internal/jcs/jcs_test.go](../go/internal/jcs/jcs_test.go) validates canonical encoding
- Test coverage includes: NFC normalization, surrogate pairs, float precision, key ordering

### C Tests
- Current status: Implementation in progress (see empty [c/src/gi.c](../c/src/gi.c))
- When adding C tests, follow the pure function pattern in [c/include/vax.h](../c/include/vax.h)

## Common Pitfalls

1. **Using standard JSON serialization** → Always use VAX-JCS canonicalization
2. **Attempting cross-actor merging** → VAX intentionally doesn't do this; detect divergence instead
3. **Backend-side normalization** → Normalize producer-side per IRP principle
4. **Assuming global ordering** → Each actor has independent linear history; no total order exists

## File Navigation

- **Spec & Philosophy:** [docs/SPECIFICATION.md](../docs/SPECIFICATION.md), [docs/ARCHITECTURE.md](../docs/ARCHITECTURE.md)
- **C Core:** 
  - [c/include/vax.h](../c/include/vax.h) — API definitions (15 functions)
  - [c/src/gi.c](../c/src/gi.c) — HMAC-SHA256 gi derivation ✅
  - [c/src/verify.c](../c/src/verify.c) — L0 verification logic ✅
  - [c/src/sai.c](../c/src/sai.c), [c/src/sae.c](../c/src/sae.c) — TODO
  - [c/CMakeLists.txt](../c/CMakeLists.txt) — Build configuration (Clang + OpenSSL)
  - [c/BUILD.md](../c/BUILD.md) — Compilation instructions
- **Go Implementation:** [go/internal/jcs/](../go/internal/jcs/) (canonicalization), [go/internal/sae/](../go/internal/sae/) (action envelope)
- **Build System:** [.vscode/settings.json](../.vscode/settings.json) — CMake + Clang integration
- **Reference Docs:** [docs/internal/](../docs/internal/) contains historical design docs

- **Build system:** CMake + Ninja (prefers Clang over GCC)
- **Dependencies:** OpenSSL (libssl, libcrypto) for HMAC-SHA256 and SHA256
- **Sanitizers:** Debug builds enable AddressSanitizer + UndefinedBehaviorSanitizer
- **Current status:** `gi.c` and `verify.c` implemented; `sai.c`, `sae.c` pending
## Language-Specific Notes

### C Implementation
- Functions prefixed `vax_` (namespace convention)
- Return `vax_result_t` enum for error handling
- Memory management: Caller must `free()` output from `vax_canonicalize_json()`
- Constants: `VAX_SAI_SIZE = 32`, `VAX_K_CHAIN_SIZE = 32`

### Go Implementation
- Module path: `vax` (see [go/go.mod](../go/go.mod))
- JCS package is self-contained; handles all JSON canonicalization
- Use `jcs.CanonicalizeJSON([]byte)` for raw input or `jcs.CanonicalizeValue(interface{})` for Go values
- SAE package [go/internal/sae/sae.go](../go/internal/sae/sae.go) builds action envelopes with timestamps

## Workflow Commands

```bash
# Go development
cd go
go test ./internal/jcs      # Test canonicalization
go test -coverprofile=coverage.out ./...  # With coverage
go run ./vax-demo           # Run demo

# C development (CMake + Clang)
cd c
cmake -B build -G Ninja -DCMAKE_BUILD_TYPE=Release  # Configure
cmake --build build                                  # Build libvax.a
cmake -B build -DCMAKE_BUILD_TYPE=Debug             # Debug with sanitizers

# Clean build
rm -rf c/build && cmake -B c/build -G Ninja
```

## When in Doubt

1. **"How should I encode this action?"** → Use VAX-JCS ([go/internal/jcs/jcs.go](../go/internal/jcs/jcs.go))
2. **"Should I merge these histories?"** → No. VAX detects divergence; resolution is higher-layer concern
3. **"Where's the consensus mechanism?"** → There isn't one. See [docs/ARCHITECTURE.md](../docs/ARCHITECTURE.md) §2
4. **"Can I modify SAE after creation?"** → No. Actions are immutable once hashed
