# VAX Go SDK

Zero-dependency tamper-evident action history.

## Installation
```bash
go get github.com/bbggbn/vax-action-history/go/pkg/vax
```

## Usage Flow

### 1. Provider: Define Schema
```go
import "vax/pkg/vax/sdto"

// Define action constraints
schema := sdto.NewSchemaBuilder().
    SetActionStringLength("username", "3", "20").
    SetActionNumberRange("amount", "0", "1000000").
    SetActionEnum("currency", []string{"USD", "EUR", "TWD"}).
    BuildSchema()

// Export as JSON for API transport
schemaJSON := schema  // map[string]any
```

### 2. Transport Schema
Share schema via API/config:
```json
{
  "username": {"type": "string", "min": "3", "max": "20"},
  "amount": {"type": "number", "min": "0", "max": "1000000"},
  "currency": {"type": "string", "enum": ["USD", "EUR", "TWD"]}
}
```

### 3. Consumer: Build Validated Actions
```go
import (
    "vax/pkg/vax/sdto"
    "vax/pkg/vax/sae"
)

// Parse received schema
schema := sdto.ParseSchema(schemaJSON)

// Build action with validation
saeBytes, err := sdto.NewAction("transfer", schema).
    Set("username", "alice").
    Set("amount", 500.0).
    Set("currency", "USD").
    Finalize()
// Returns canonical SAE bytes
```

### 4. Service Init: Setup Chain
```go
import "vax/pkg/vax"

// On service connect/session start
kChain := []byte("your-32-byte-secret-key-here")
genesisSAI := vax.ComputeGenesisSAI(kChain)

// Store: actor_id -> {genesisSAI, counter: 0}
```

### 5. Record Action: Compute SAI
```go
// Increment counter
counter := lastCounter + 1

// Compute gi_n
gi := vax.ComputeGI(kChain, counter)

// Compute SAI_n
sai := vax.ComputeSAI(prevSAI, saeBytes, gi)

// Store action: {sai, saeBytes, counter, timestamp}
// Update state: prevSAI = sai, lastCounter = counter
```

## Pipeline Summary
```
SchemaBuilder → JSON/API → ParseSchema → FluentAction → SAE
                                                          ↓
Service Init: ComputeGenesisSAI(k) → SAI_0
                                       ↓
Record Loop:  ComputeGI(k, n) ─┐
              prevSAI ──────────┼→ ComputeSAI → SAI_n
              SAE ──────────────┘
```

## Key Rules
- Always use `jcs.Marshal()`, never `json.Marshal()`
- Counter is big-endian uint16
- SAI chain: each action references `prevSAI`
- Validation happens at `.Set()`, errors on `.Finalize()`

## Testing
```bash
go test ./pkg/vax/...
```

## Docs
- [Changelog](doct/changelog.md)
- [Specification](../docs/SPECIFICATION.md)
- [Architecture](../docs/ARCHITECTURE.md)
