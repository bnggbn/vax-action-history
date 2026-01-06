# VAX Copilot Instructions

Git-like tamper-evident action history. Multi-language SDK (C/Go/TS) with deterministic output.

## Structure
```
go/pkg/vax/     # Public SDK (zero internal deps)
├── jcs/        # JSON Canonicalization
├── sae/        # Semantic Action Envelope
├── sdto/       # Schema-Driven Type Objects
└── vax.go      # L0 crypto

ts/src/         # TypeScript port
c/src/          # Pure crypto (OpenSSL)
```

## Critical Rules

**1. Always use JCS, never `json.Marshal`**
```go
import "vax/pkg/vax/jcs"
saeBytes, err := jcs.Marshal(action)  // ✓ deterministic
```

**2. `pkg/vax/` = public SDK** → no `internal/` imports

**3. Language boundaries**
- C: crypto only (`vax_compute_gi`, `vax_compute_sai`, `vax_verify_action`)
- Go/TS: JSON, state, validation

**4. Hash formula**
```
SAI_n = SHA256("VAX-SAI" || prevSAI || SHA256(SAE) || gi_n)
gi_n  = HMAC_SHA256(k_chain, "VAX-GI" || counter)  # big-endian uint16
```

## SDTO Pattern
Provider defines schema → consumer builds + validates → SAE bytes

```go
// Provider
schema := sdto.NewSchemaBuilder().
    SetActionStringLength("name", "1", "50").
    SetActionNumberRange("amount", "0", "1000000").
    BuildSchema()

// Consumer
saeBytes, _ := sdto.NewAction("transfer", schema).
    Set("name", "Alice").
    Set("amount", 500.0).
    Finalize()
```

Validation at `.Set()` time, errors on `.Finalize()`.

## Test Commands
```bash
go test ./...                   # All Go tests
go test ./pkg/vax/...          # SDK only
cd c && cmake -B build && cmake --build build && ctest --test-dir build
cd ts && npm test              # TypeScript
```

## Common Pitfalls
- `json.Marshal` → use `jcs.Marshal`
- Counter → big-endian uint16
- String min/max → `strconv.Atoi`
- `pkg/` → cannot import `internal/`

## Docs
- [SPECIFICATION.md](docs/SPECIFICATION.md)
- [go/doct/changelog.md](go/doct/changelog.md)
- [ts/CHANGELOG.md](ts/CHANGELOG.md)
- [test-vectors.json](test-vectors.json) — cross-lang verification

