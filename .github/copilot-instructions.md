# VAX Copilot Instructions

Git-like tamper-evident action history. Multi-language SDK (C/Go/TS) with deterministic output.

## Structure
```
go/pkg/vax/     # Public SDK (zero internal deps)
├── jcs/        # JSON Canonicalization
├── sae/        # Semantic Action Envelope
├── sdto/       # Schema-Driven Type Objects
└── vax.go      # L0 crypto

ts/src/         # TypeScript port (mirrors Go structure)
├── jcs/        # JSON Canonicalization
├── sae/        # Semantic Action Envelope
├── sdto/       # Schema-Driven Type Objects
└── vax.ts      # L0 crypto

c/src/          # Pure crypto (OpenSSL)
```

## Critical Rules

**1. Always use JCS, never `json.Marshal` or `JSON.stringify`**
```go
import "vax/pkg/vax/jcs"
saeBytes, err := jcs.Marshal(action)  // ✓ deterministic
```
```typescript
import { marshal } from './jcs';
const saeBytes = marshal(action);  // ✓ deterministic
```

**2. `pkg/vax/` = public SDK** → no `internal/` imports

**3. Language boundaries**
- C: crypto only (`vax_compute_gi`, `vax_compute_sai`, `vax_verify_action`)
- Go/TS: JSON, state, validation
- **Architecture sync**: Go is reference, TS mirrors structure

**4. Hash formula (v0.7 - deterministic)**
```
SAI_n = SHA256("VAX-SAI" || prevSAI || SHA256(SAE))
Genesis: SAI_0 = SHA256("VAX-GENESIS" || actor_id || genesis_salt)
```

## SDTO Pattern
Provider defines schema → consumer builds + validates → SAE bytes

```go
// Provider
schema := sdto.NewSchemaBuilder().
    SetActionStringLength("name", "1", "50").
    SetActionNumberRange("amount", "0", "1000000").
    SetActionSign("signature", "ed25519").  // NEW: signature support
    BuildSchema()

// Consumer
saeBytes, _ := sdto.NewAction("transfer", schema).
    Set("name", "Alice").
    Set("amount", 500.0).
    Set("signature", "...").
    Finalize()  // Returns JCS-canonicalized SAE bytes
```

```typescript
// TypeScript (identical pattern)
const schema = newSchemaBuilder()
    .setActionStringLength("name", "1", "50")
    .setActionNumberRange("amount", "0", "1000000")
    .setActionSign("signature", "ed25519")
    .buildSchema();

const saeBytes = newAction("transfer", schema)
    .set("name", "Alice")
    .set("amount", 500.0)
    .set("signature", "...")
    .finalize();  // Returns Buffer (JCS bytes)
```

**Key**: Validation at `.Set()` time, errors on `.Finalize()`.

## Cross-Language Verification

Use `test-vectors.json` for consistency:
```bash
# All implementations must produce identical outputs
go test ./pkg/vax/jcs/jcs_cross_lang_test.go
npm test -- jcs.cross-lang.test.ts
```

## Test Commands
```bash
go test ./...                   # All Go tests
go test ./pkg/vax/...          # SDK only
cd ts && npm install && npm test  # TypeScript
cd c && cmake -B build && cmake --build build && ctest --test-dir build
```

## Recent Changes (2026-01-07)

### TypeScript Refactor
- Added `src/sae/` module (matches Go's `pkg/vax/sae/`)
- `FluentAction.finalize()` now returns `Buffer` (SAE bytes), not object
- `computeSAI()` is now deterministic (no random gi)
- Added `verifyAction()` full verification (crypto + schema)
- Renamed old `verifyAction()` → `verifyPrevSAI()`

### SDTO Enhancements (Both Languages)
- New `sign` type for signature fields
- `SchemaBuilder.SetActionSign()` / `setActionSign()`
- `SchemaBuilder.SetActionSignMulti()` / `setActionSignMulti()`
- Server-side `ValidateData()` / `validateData()` function

## Common Pitfalls
- `json.Marshal` / `JSON.stringify` → use JCS `marshal()`
- String min/max → numeric parsing (`strconv.Atoi` / `parseInt`)
- `pkg/` → cannot import `internal/`
- TS: `Finalize()` returns bytes now, not `{ actionType, data }`
- Cross-lang: Always verify with test vectors

## Docs
- [SPECIFICATION.md](docs/SPECIFICATION.md) — Technical spec
- [ARCHITECTURE.md](docs/ARCHITECTURE.md) — Design philosophy
- [go/doct/changelog.md](go/doct/changelog.md) — Go changes
- [ts/CHANGELOG.md](ts/CHANGELOG.md) — TypeScript changes
- [test-vectors.json](test-vectors.json) — Cross-lang verification

