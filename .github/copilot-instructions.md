# VAX Copilot Instructions

Git-like tamper-evident action history. Multi-language SDK (C/Go/TS) with deterministic cryptographic output.

## Architecture Overview

**Design Philosophy:** VAX is a **tool, not a protocol** — provides cryptographic primitives without enforcing storage, authorization, or transport decisions. Like Git guarantees code history integrity, VAX guarantees action history integrity.

**Core Layers:**
- **L0 (Crypto)**: Hash chain computation, verification (C reference, Go/TS ports)
- **SDTO (Schema)**: Provider defines rules, consumer builds + validates
- **JCS (Canonicalization)**: Deterministic JSON encoding for reproducible hashes

**Language Responsibilities:**
- **C (`c/src/`)**: Frozen reference for crypto primitives only (`vax_compute_sai`, `vax_verify_action`)
- **Go (`go/pkg/vax/`)**: Public SDK with zero `internal/` deps — reference implementation
- **TypeScript (`ts/src/`)**: Mirrors Go structure exactly (JCS, SAE, SDTO, vax.ts)

```
go/pkg/vax/     ts/src/         Purpose
├── jcs/        ├── jcs/        JSON canonicalization (VAX-JCS)
├── sae/        ├── sae/        Semantic Action Envelope
├── sdto/       ├── sdto/       Schema-Driven Type Objects
└── vax.go      └── vax.ts      L0 crypto (SAI computation/verification)
```

## Critical Development Rules

### 1. Always Use JCS for Serialization
**Never** use standard JSON serialization. JCS ensures deterministic byte output.

```go
// ✓ Correct
import "vax/pkg/vax/jcs"
saeBytes, err := jcs.Marshal(action)

// ✗ Wrong - non-deterministic
jsonBytes, _ := json.Marshal(action)
```

```typescript
// ✓ Correct
import { marshal } from './jcs';
const saeBytes = marshal(action);

// ✗ Wrong - non-deterministic
const jsonBytes = JSON.stringify(action);
```

### 2. No Internal Dependencies in `pkg/vax/`
Go's `pkg/vax/` is the public SDK — cannot import from `internal/`. All functionality must be self-contained.

### 3. Hash Formulas (v0.7 Deterministic)
```
SAI_n = SHA256("VAX-SAI" || prevSAI || SHA256(SAE))
Genesis: SAI_0 = SHA256("VAX-GENESIS" || actor_id || genesis_salt)
```
**No random `gi`** in v0.7 — inputs produce identical outputs.

### 4. Architecture Sync Across Languages
- Go is the **reference implementation**
- TypeScript **mirrors** Go structure and behavior
- Use `test-vectors.json` to verify cross-language consistency
- All implementations must produce **byte-identical** outputs for same inputs

## SDTO Pattern (Schema-Driven Type Objects)

Provider-consumer pattern: schema provider defines rules, consumer builds + validates, outputs canonical bytes.

```go
// Provider defines schema
schema := sdto.NewSchemaBuilder().
    SetActionStringLength("name", "1", "50").
    SetActionNumberRange("amount", "0", "1000000").
    SetActionEnum("status", []string{"pending", "completed"}).
    BuildSchema()

// Consumer builds action (validates immediately on .Set())
saeBytes, err := sdto.NewAction("transfer", schema).
    Set("name", "Alice").
    Set("amount", 500.0).
    Set("status", "pending").
    Finalize()  // Returns JCS-canonicalized SAE bytes (Buffer/[]byte)
```

**Key Behaviors:**
- Validation happens **at `.Set()` time** — fail fast
- `.Finalize()` returns **bytes** (Buffer in TS, []byte in Go), not objects
- String min/max use **numeric parsing** (`strconv.Atoi` / `parseInt`)
- Errors accumulate and throw on finalize

## Developer Workflows

### Testing Strategy
```bash
# Go: Run all tests including cross-language verification
go test ./...                              # All tests
go test ./pkg/vax/...                      # SDK only
go test ./pkg/vax/jcs/jcs_cross_lang_test.go  # Cross-lang JCS

# TypeScript: Mirror Go test structure
cd ts && npm install && npm test           # All tests
npm test -- jcs.cross-lang.test.ts         # Cross-lang JCS

# C: Reference crypto verification
cd c && cmake -B build && cmake --build build && ctest --test-dir build
```

### Cross-Language Verification
`test-vectors.json` contains canonical inputs/outputs for all languages:
- Same input JSON → same JCS bytes → same SAI
- Run cross-lang tests after any JCS/crypto changes
- Add new vectors when introducing edge cases

### Building SAE (Semantic Action Encoding)
```go
import "vax/pkg/vax/sae"

envelope := sae.Envelope{
    ActionType: "transfer",
    SDTO: map[string]interface{}{
        "from": "alice",
        "to": "bob",
        "amount": 100.0,
    },
}
saeBytes, _ := jcs.Marshal(envelope)  // Must use JCS
```

### Full Verification Flow (Server-Side)
```go
import "vax/pkg/vax"

// Backend verifies: prevSAI chain + schema compliance + SAI computation
envelope, err := vax.VerifyAction(
    expectedPrevSAI,
    clientPrevSAI,
    saeBytes,           // Already JCS-marshaled by client
    clientSAI,
    schema,
)
```

## Common Pitfalls & Solutions

| ❌ Mistake | ✓ Correct |
|-----------|-----------|
| `json.Marshal(obj)` | `jcs.Marshal(obj)` |
| `JSON.stringify(obj)` | `marshal(obj)` from `./jcs` |
| String "10" < "2" (lexical) | Parse with `strconv.Atoi` / `parseInt` |
| Import `internal/` from `pkg/` | Keep `pkg/vax/` self-contained |
| `finalize()` returns object | Returns `Buffer` (TS) / `[]byte` (Go) |
| Expect random `gi` in SAI | v0.7 is deterministic (no gi) |

## Key Files & Patterns

**Must-read documentation:**
- [`docs/SPECIFICATION.md`](docs/SPECIFICATION.md) — Byte-level behavior (SAI formulas, JCS rules)
- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) — Design philosophy ("tool not protocol")
- [`test-vectors.json`](test-vectors.json) — Cross-language verification oracle

**Example patterns:**
- [`go/pkg/vax/sdto/integration_test.go`](go/pkg/vax/sdto/integration_test.go) — End-to-end SDTO flow
- [`go/pkg/vax/jcs/jcs_cross_lang_test.go`](go/pkg/vax/jcs/jcs_cross_lang_test.go) — Cross-lang verification
- [`ts/src/vax.ts`](ts/src/vax.ts) — TypeScript API (mirrors Go)

**Changelogs for recent changes:**
- [`go/doct/changelog.md`](go/doct/changelog.md) — Go SDK evolution
- [`ts/CHANGELOG.md`](ts/CHANGELOG.md) — TypeScript breaking changes

## Recent Breaking Changes (2026-01-07)

- **TypeScript**: `FluentAction.finalize()` now returns `Buffer` (SAE bytes), not `{ actionType, data }`
- **Both**: `computeSAI()` is deterministic (removed random gi generation)
- **Both**: Added `verifyAction()` for full crypto + schema verification
- **SDTO**: String length validation uses numeric parsing (not lexical comparison)

## Non-Goals (By Design)

VAX intentionally does **not** provide:
- ❌ Signature mechanism (use standard crypto libraries)
- ❌ Storage/database structure (define your own)
- ❌ Authorization logic (implement at L1/L2)
- ❌ Transport protocol (HTTP/gRPC/WebSocket — your choice)
- ❌ Key management (use your KMS)
- ❌ Cross-actor consensus (single-actor linear history only)

