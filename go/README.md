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
// Returns canonical SAE bytes (unsigned)
```

### 4. Client: Compute SAI
```go
import (
    "vax/pkg/vax"
    "crypto/rand"
)

// First action: Compute genesis SAI
actorID := "user123:device456"
genesisSalt := make([]byte, 16)
rand.Read(genesisSalt)

genesisSAI, err := vax.ComputeGenesisSAI(actorID, genesisSalt)
if err != nil {
    // handle error
}

// For subsequent actions: Compute SAI from previous
prevSAI := genesisSAI  // or lastSAI from storage
sai, err := vax.ComputeSAI(prevSAI, saeBytes)
if err != nil {
    // handle error
}

// Submit to backend: {sai, saeBytes, prevSAI}
```

### 5. Backend: Verify and Sign
```go
import (
    "vax/pkg/vax"
    "crypto/ed25519"
)

// Backend receives: {expectedPrevSAI, prevSAI, saeBytes, clientSAI}
// Backend has: schema, Ed25519 privateKey

signedEnvelope, err := vax.VerifyAction(
    expectedPrevSAI,  // What backend expects
    prevSAI,          // What client claims
    saeBytes,         // Client's SAE
    clientSAI,        // Client's computed SAI
    schema,           // Backend's schema
    privateKey,       // Backend's signing key
)

if err != nil {
    // Verification failed: reject action
    // Possible errors:
    // - ErrInvalidPrevSAI: Chain broken
    // - ErrSAIMismatch: SAI computation wrong
    // - Schema validation failed
}

// Success: signedEnvelope contains backend signature
// Store: {sai, signedEnvelope, timestamp}
// Update state: prevSAI = sai
```

## Pipeline Summary
```
Provider Side:
  SchemaBuilder → JSON Schema → Transport to Consumer

Consumer Side:
  Schema → FluentAction.Set() → Finalize() → SAE (unsigned)
                                               ↓
  actorID + genesisSalt → ComputeGenesisSAI → SAI_0 (first action)
                                               ↓
  prevSAI + SAE ──────────→ ComputeSAI ─────→ SAI_n (subsequent actions)
                                               ↓
                                        Submit to Backend

Backend Side:
  Verify prevSAI continuity
      ↓
  Verify SAI = SHA256("VAX-SAI" || prevSAI || SHA256(SAE))
      ↓
  Verify SDTO against Schema
      ↓
  Sign SAE with Ed25519 (action enters history)
      ↓
  Store and update prevSAI
```

## Core Formulas

```go
// Genesis (first action for an Actor)
SAI_0 = SHA256("VAX-GENESIS" || actorID || genesisSalt)

// Subsequent actions
SAI_n = SHA256("VAX-SAI" || prevSAI || SHA256(SAE))

// Where:
// - actorID: "user_id:device_id"
// - genesisSalt: 16 random bytes, persistent per Actor
// - prevSAI: previous action's SAI (or SAI_0 for first action)
// - SAE: Semantic Action Encoding (canonical JSON)
```

## Key Principles

### Client-Side
- Always use `jcs.Marshal()`, never `json.Marshal()`
- Validation happens at `.Set()` time, errors collected at `.Finalize()`
- SAI chain: each action references `prevSAI`
- Client computes SAI before submitting to backend

### Backend-Side
- Backend **verifies** SAI, never computes it
- Backend **signs** SAE to mark "action entered history"
- Schema is backend's authority: frontend changes don't affect backend
- Backend never repairs or modifies client data (IRP principle)

### Security Properties
- **Append-only**: prevSAI chain prevents reordering
- **Tamper-evident**: SAI changes if SAE changes
- **Backend authority**: Backend signature = official record
- **Schema enforcement**: Both sides validate, backend is source of truth

## Testing
```bash
go test ./pkg/vax/...
```

## API Reference

### Core Functions

```go
// Compute genesis SAI for an Actor
func ComputeGenesisSAI(actorID string, genesisSalt []byte) ([]byte, error)

// Compute SAI for an action
func ComputeSAI(prevSAI, saeBytes []byte) ([]byte, error)

// Backend verification: verify and sign action
func VerifyAction(
    expectedPrevSAI []byte,
    prevSAI []byte,
    saeBytes []byte,
    clientProvidedSAI []byte,
    schema map[string]sdto.FieldSpec,
    privateKey ed25519.PrivateKey,
) (*sae.Envelope, error)
```

### Schema Builder

```go
// Create schema
builder := sdto.NewSchemaBuilder()
builder.SetActionStringLength("field", "min", "max")
builder.SetActionNumberRange("field", "min", "max")
builder.SetActionEnum("field", []string{"val1", "val2"})
builder.SetActionSign("field", "ed25519")        // Require client signature
builder.SetActionSignMulti("field", []string{"ed25519", "rsa"})
schema := builder.BuildSchema()

// Build action from schema
saeBytes, err := sdto.NewAction("actionType", schema).
    Set("field1", value1).
    Set("field2", value2).
    Finalize()
```

## Error Codes

```go
var (
    ErrInvalidCounter  = errors.New("invalid counter")     // Reserved
    ErrInvalidPrevSAI  = errors.New("invalid prevSAI")     // Chain broken
    ErrSAIMismatch     = errors.New("SAI mismatch")        // SAI verification failed
    ErrOutOfMemory     = errors.New("out of memory")       // Memory allocation failed
    ErrInvalidInput    = errors.New("invalid input")       // Invalid parameters
    ErrCounterOverflow = errors.New("counter overflow")    // Reserved
)
```

## Docs
- [Changelog](doct/changelog.md)
- [L0 Specification](../docs/SPECIFICATION.md)
- [Architecture & Design Philosophy](../docs/ARCHITECTURE.md)


## License
MIT

## Contributing
See [CONTRIBUTING.md](CONTRIBUTING.md)
```

---
