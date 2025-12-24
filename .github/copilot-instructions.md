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

### 4. C Core = Pure Crypto Only (Critical!)
**The C implementation MUST NOT handle JSON canonicalization or state management.** Only cryptographic primitives:
- ✅ `vax_compute_gi()`, `vax_compute_sai()`, `vax_verify_action()` 
- ❌ No `vax_chain_*` API (state management belongs in Go/TS)
- ❌ No `vax_is_canonical()` or JSON handling

See conversation context for the rationale: C core stays pure, Go server handles JSON+state, JS/TS frontend handles canonicalization.

## Critical Implementation Rules

### Canonical JSON (VAX-JCS)
**NEVER use `json.Marshal()` or standard JSON encoding for action data.**

- **Go:** Use [go/internal/jcs/jcs.go](../go/internal/jcs/jcs.go) - `jcs.Marshal()` or `jcs.CanonicalizeJSON()`
- **C:** JSON canonicalization is OUT OF SCOPE for C core (done by caller)
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

### Hash Chain Construction (Two-Stage Hash)
Actions use a **two-stage hash** to avoid malloc and improve security:

```
sae_hash = SHA256(SAE)
SAI_n = SHA256("VAX-SAI" || prevSAI || sae_hash || gi_n)
```

- `gi_n`: per-action entropy via `HMAC_SHA256(k_chain, "VAX-GI" || counter)` (big-endian)
- `prevSAI`: previous action's hash (32 bytes)
- `sae_hash`: SHA256 of canonical JSON bytes

**Never compute hashes manually**—use:
- C: `vax_compute_gi()`, `vax_compute_sai()` in [c/include/vax.h](../c/include/vax.h)
- Go: Implement following the C reference (not yet in Go codebase)

### State Management
**C core provides ONLY pure functions** (no state):
- `vax_compute_gi()` — HMAC-SHA256 gi derivation
- `vax_compute_sai()` — Two-stage hash SAI computation  
- `vax_compute_genesis_sai()` — Genesis SAI with actor_id
- `vax_verify_action()` — Crypto-only verification (no JSON validation)

**State management (counter, prevSAI tracking) belongs in:**
- Go server layer (not yet implemented)
- TypeScript/JavaScript frontend (future work)

## Testing Strategy

### Go Tests
```bash
cd go
go test ./...                              # Run all tests
go test ./internal/jcs                     # JCS canonicalization only
go test -coverprofile=coverage.out ./...   # With coverage
go run ./vax-demo                          # Run demo
```
- Key test: [go/internal/jcs/jcs_test.go](../go/internal/jcs/jcs_test.go) validates canonical encoding
- Test coverage includes: NFC normalization, surrogate pairs, float precision, key ordering

### C Tests (CMake + CTest)
```bash
cd c
cmake -B build -G Ninja -DCMAKE_BUILD_TYPE=Debug   # Configure with sanitizers
cmake --build build                                 # Build library + tests
ctest --test-dir build --output-on-failure          # Run all tests
./build/test_gi                                     # Run specific test
```
- **Test files:** [c/test/test_gi.c](../c/test/test_gi.c) (8 tests), [c/test/test_sai.c](../c/test/test_sai.c) (7 tests), [c/test/test_verify.c](../c/test/test_verify.c) (5 tests)
- **Test helpers:** [c/test/test_common.h](../c/test/test_common.h) — `print_hex()`, `compare_bytes()`
- **Sanitizers:** AddressSanitizer + UndefinedBehaviorSanitizer enabled in Debug builds
- **Test vectors verified with OpenSSL:** See test file comments for `openssl dgst` commands

## Common Pitfalls

1. **Using standard JSON serialization** → Always use VAX-JCS canonicalization (Go: [go/internal/jcs/](../go/internal/jcs/))
2. **Adding state management to C core** → C is pure crypto only; state belongs in Go/TS
3. **Adding JSON handling to C** → C core does NOT validate or canonicalize JSON
4. **Wrong hash formula** → Use two-stage hash: `SHA256("VAX-SAI" || prevSAI || SHA256(SAE) || gi)`
5. **Attempting cross-actor merging** → VAX intentionally doesn't do this; detect divergence instead
6. **Backend-side normalization** → Normalize producer-side per IRP principle
7. **Assuming global ordering** → Each actor has independent linear history; no total order exists
8. **Counter encoding** → Must be big-endian uint16 in gi computation

## File Navigation

- **Spec & Philosophy:** [docs/SPECIFICATION.md](../docs/SPECIFICATION.md), [docs/ARCHITECTURE.md](../docs/ARCHITECTURE.md)
- **C Core (Pure Crypto):** 
  - [c/include/vax.h](../c/include/vax.h) — API definitions (8 functions, 7 error codes, 4 constants)
  - [c/src/gi.c](../c/src/gi.c) — HMAC-SHA256 gi derivation ✅
  - [c/src/sai.c](../c/src/sai.c) — Two-stage hash SAI computation ✅
  - [c/src/verify.c](../c/src/verify.c) — L0 verification logic (crypto only) ✅
  - [c/test/](../c/test/) — Test suite (test_gi.c, test_sai.c, test_verify.c) ✅
  - [c/CMakeLists.txt](../c/CMakeLists.txt) — Build configuration (Clang + Ninja + OpenSSL)
  - [c/BUILD.md](../c/BUILD.md) — Compilation instructions
- **Go Implementation:** 
  - [go/internal/jcs/](../go/internal/jcs/) — Canonical JSON (VAX-JCS)
  - [go/internal/sae/](../go/internal/sae/) — Action envelope (timestamp, actor metadata)
  - [go/vax-demo/](../go/vax-demo/) — Demo application
- **Build System:** 
  - [.vscode/settings.json](../.vscode/settings.json) — CMake + Clang integration
  - [.vscode/extensions.json](../.vscode/extensions.json) — Recommended extensions (C/C++, CMake Tools)
- **Reference Docs:** [docs/internal/](../docs/internal/) contains historical design docs

## Build System (C)

**Tool Chain:**
- **Build system:** CMake 3.15+ (NOT a compiler—it's a build config generator)
- **Generator:** Ninja (fast incremental builds)
- **Compiler:** Clang/LLVM (auto-detected, falls back to GCC)
- **Dependencies:** OpenSSL (libssl, libcrypto) for HMAC-SHA256 and SHA256
- **Standard:** C11
- **Sanitizers:** AddressSanitizer + UndefinedBehaviorSanitizer (Debug builds only)

**Build Workflow:**
```bash
cd c

# Configure (first time or after CMakeLists.txt changes)
cmake -B build -G Ninja -DCMAKE_BUILD_TYPE=Release

# Build
cmake --build build         # Output: build/libvax.a

# Run tests
ctest --test-dir build --output-on-failure

# Clean rebuild
rm -rf build && cmake -B build -G Ninja
```

**Debug vs Release:**
```bash
# Debug: -O0 + sanitizers (AddressSanitizer, UndefinedBehaviorSanitizer)
cmake -B build -G Ninja -DCMAKE_BUILD_TYPE=Debug

# Release: -O3 + LTO (Link-Time Optimization)
cmake -B build -G Ninja -DCMAKE_BUILD_TYPE=Release
```
## Language-Specific Notes

### C Implementation
- **Namespace:** Functions prefixed `vax_` 
- **Error handling:** Return `vax_result_t` enum (`VAX_OK`, `VAX_ERR_*`)
- **Memory:** All functions use caller-provided buffers (no malloc/free in crypto primitives)
- **Constants:** `VAX_SAI_SIZE = 32`, `VAX_K_CHAIN_SIZE = 32`, `VAX_GI_SIZE = 32`, `VAX_GENESIS_SALT_SIZE = 16`
- **Endianness:** Counter MUST be big-endian in gi computation (see [c/src/gi.c](../c/src/gi.c))

### Go Implementation
- **Module path:** `vax` (see [go/go.mod](../go/go.mod))
- **JCS package:** Self-contained; handles all JSON canonicalization
  - `jcs.Marshal(interface{})` → `[]byte` (canonical JSON)
  - `jcs.CanonicalizeJSON([]byte)` → `[]byte` (re-canonicalize)
- **SAE package:** [go/internal/sae/sae.go](../go/internal/sae/sae.go) builds action envelopes with timestamps
- **No CGo yet:** Go implementation will follow C reference without direct linking

## Workflow Commands

```bash
# Go development
cd go
go test ./internal/jcs                           # Test canonicalization only
go test -coverprofile=coverage.out ./...         # All tests with coverage
go run ./vax-demo                                # Run demo

# C development (CMake + Clang + Ninja)
cd c
cmake -B build -G Ninja -DCMAKE_BUILD_TYPE=Release  # Configure
cmake --build build                                  # Build libvax.a + tests
ctest --test-dir build --output-on-failure           # Run all tests
cmake -B build -DCMAKE_BUILD_TYPE=Debug             # Debug with sanitizers

# Clean build
rm -rf c/build && cmake -B c/build -G Ninja

# Verify test vectors manually (OpenSSL)
# See test file comments for exact commands, example:
printf '\x56\x41\x58\x2d\x47\x49\x00\x01' | openssl dgst -sha256 -hmac "..."
```

## When in Doubt

1. **"How should I encode this action?"** → Use VAX-JCS ([go/internal/jcs/jcs.go](../go/internal/jcs/jcs.go))
2. **"Should I add JSON handling to C?"** → No. C core = crypto only. JSON is Go/TS responsibility.
3. **"Should I merge these histories?"** → No. VAX detects divergence; resolution is higher-layer concern.
4. **"Where's the consensus mechanism?"** → There isn't one. See [docs/ARCHITECTURE.md](../docs/ARCHITECTURE.md) §2.
5. **"Can I modify SAE after creation?"** → No. Actions are immutable once hashed.
6. **"Should I add `vax_chain_*` functions?"** → No. State management belongs in Go/TS, not C core.
