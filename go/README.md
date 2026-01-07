# VAX Go SDK

Pure Go implementation of VAX cryptographic primitives.

Zero dependencies. Deterministic output.

---

## Installation

```bash
go get github.com/bnggbn/vax-action-history/go/pkg/vax
```

---

## Philosophy

VAX provides **primitives**, not a complete system.

**You control:**
- Storage structure
- Whether to add signatures
- Authorization logic
- Transport protocol

**VAX provides:**
- SAI chain computation
- JCS canonicalization
- Schema validation

---

## Quick Start

```go
package main

import (
    "fmt"
    "vax/pkg/vax"
)

func main() {
    // 1. Genesis
    actorID := "user123:device456"
    genesisSalt := vax.GenerateGenesisSalt()
    genesisSAI, _ := vax.ComputeGenesisSAI(actorID, genesisSalt)

    // 2. Build action
    saeBytes := buildAction("transfer", data)

    // 3. Compute SAI
    sai, _ := vax.ComputeSAI(genesisSAI, saeBytes)

    // 4. Store (your structure)
    store(sai, saeBytes, genesisSAI)
}
```

---

## Core API

### Genesis

```go
func ComputeGenesisSAI(actorID string, genesisSalt []byte) ([]byte, error)
```

Compute genesis SAI for an Actor.

**Formula:**
```
SAI_0 = SHA256("VAX-GENESIS" || actorID || genesisSalt)
```

**Example:**
```go
actorID := "user123:device456"
salt := vax.GenerateGenesisSalt()  // 16 random bytes
sai, err := vax.ComputeGenesisSAI(actorID, salt)
```

---

### Chain

```go
func ComputeSAI(prevSAI, saeBytes []byte) ([]byte, error)
```

Compute SAI for an action.

**Formula:**
```
SAI_n = SHA256("VAX-SAI" || prevSAI || SHA256(SAE))
```

**Example:**
```go
sae := buildSAE("transfer", data)
sai, err := vax.ComputeSAI(prevSAI, sae)
```

---

### Verification

```go
func VerifyChain(expectedPrevSAI, saeBytes, clientSAI []byte) error
```

Verify SAI chain integrity.

**Checks:**
1. prevSAI continuity
2. SAI computation correctness

**Example:**
```go
err := vax.VerifyChain(expectedPrevSAI, saeBytes, clientSAI)
if err != nil {
    // Chain broken or SAI mismatch
}
```

---

## Schema-Driven Validation (SDTO)

### Define Schema

```go
import "vax/pkg/vax/sdto"

schema := sdto.NewSchemaBuilder().
    SetActionStringLength("username", "3", "20").
    SetActionNumberRange("amount", "0", "1000000").
    SetActionEnum("currency", []string{"USD", "EUR", "TWD"}).
    BuildSchema()
```

### Build Validated Action

```go
saeBytes, err := sdto.NewAction("transfer", schema).
    Set("username", "alice").
    Set("amount", 500.0).
    Set("currency", "USD").
    Finalize()  // Returns canonical JSON bytes
```

**Validation happens at `.Set()` time.** Errors collected and returned at `.Finalize()`.

### Server-Side Validation

```go
// Backend validates SDTO against schema
err := sdto.ValidateData(action.SDTO, schema)
if err != nil {
    // Schema violation
}
```

---

## JCS (JSON Canonicalization)

```go
import "vax/pkg/vax/jcs"

// Marshal to canonical JSON
canonical := jcs.Marshal(obj)

// Always produces identical output
obj1 := map[string]any{"b": 2, "a": 1}
obj2 := map[string]any{"a": 1, "b": 2}

bytes1 := jcs.Marshal(obj1)  // {"a":1,"b":2}
bytes2 := jcs.Marshal(obj2)  // {"a":1,"b":2}
// bytes1 == bytes2 ✅
```

**Never use `json.Marshal()` for SAE. Always use `jcs.Marshal()`.**

---

## SAE (Semantic Action Envelope)

```go
import "vax/pkg/vax/sae"

saeBytes := sae.BuildSAE("transfer", map[string]any{
    "username": "alice",
    "amount":   500.0,
    "currency": "USD",
})

// Returns canonical JSON:
// {"action_type":"transfer","sdto":{...},"timestamp":1704672000000}
```

**Structure:**
```go
type Envelope struct {
    ActionType string         `json:"action_type"`
    Timestamp  int64          `json:"timestamp"`
    SDTO       map[string]any `json:"sdto"`
}
```

**No signature field.** If you need signatures, add them yourself.

---

## Complete Workflow

### Client Side

```go
import (
    "vax/pkg/vax"
    "vax/pkg/vax/sdto"
)

// 1. Get schema from backend
schema := backend.GetSchema("transfer")

// 2. Build validated action
saeBytes, err := sdto.NewAction("transfer", schema).
    Set("username", "alice").
    Set("amount", 500.0).
    Set("currency", "USD").
    Finalize()

// 3. Compute SAI
prevSAI := getLastSAI()
sai, err := vax.ComputeSAI(prevSAI, saeBytes)

// 4. Submit to backend
backend.Submit(Request{
    SAI:     sai,
    SAE:     saeBytes,
    PrevSAI: prevSAI,
})
```

### Backend Side

```go
// 1. Verify chain
err := vax.VerifyChain(expectedPrevSAI, req.SAE, req.SAI)
if err != nil {
    return err
}

// 2. Validate schema
var env sae.Envelope
json.Unmarshal(req.SAE, &env)
err = sdto.ValidateData(env.SDTO, schema)
if err != nil {
    return err
}

// 3. Optional: Add signature
signature := ed25519.Sign(privateKey, req.SAE)

// 4. Store (your structure)
db.Store(ActionRecord{
    SAI:       req.SAI,
    SAE:       req.SAE,
    PrevSAI:   req.PrevSAI,
    Signature: signature,  // optional
    Timestamp: time.Now(),
})
```

---

## Optional: Signatures

VAX doesn't handle signatures. Use Go standard library:

```go
import "crypto/ed25519"

// Generate key pair
publicKey, privateKey, _ := ed25519.GenerateKey(nil)

// Sign SAE
signature := ed25519.Sign(privateKey, saeBytes)

// Verify
valid := ed25519.Verify(publicKey, saeBytes, signature)
```

**Storage is your choice:**

```go
// Option A: Separate field
type ActionRecord struct {
    SAI       []byte
    SAE       []byte
    PrevSAI   []byte
    Signature []byte  // Independent
}

// Option B: Wrap in metadata
type ActionWithMeta struct {
    Action struct {
        SAI     []byte
        SAE     []byte
        PrevSAI []byte
    }
    Metadata struct {
        Signature []byte
        Timestamp time.Time
        Author    string
    }
}

// Your choice!
```

---

## Helpers

```go
// Generate random genesis salt
salt := vax.GenerateGenesisSalt()  // 16 bytes

// Hex encoding
hex := vax.ToHex(sai)
sai, err := vax.FromHex(hex)

// Constants
vax.SAI_SIZE           // 32
vax.GENESIS_SALT_SIZE  // 16
```

---

## Error Handling

```go
var (
    ErrInvalidInput    = errors.New("invalid input")
    ErrInvalidPrevSAI  = errors.New("invalid prevSAI")
    ErrSAIMismatch     = errors.New("SAI mismatch")
    ErrCounterOverflow = errors.New("counter overflow")  // Reserved
)
```

---

## Testing

```bash
go test ./pkg/vax/...
go test ./pkg/vax/... -v
go test ./pkg/vax/... -cover
```

---

## Examples

See [examples/](examples/) for complete workflows:

- **[minimal](examples/minimal/)** - Simplest usage, no signatures
- **[with-signature](examples/with-signature/)** - Add backend signature
- **[multi-actor](examples/multi-actor/)** - Multiple actors
- **[verification](examples/verification/)** - Audit and verification

---

## Key Principles

### 1. Always use JCS

```go
// ✅ Correct
import "vax/pkg/vax/jcs"
bytes := jcs.Marshal(obj)

// ❌ Wrong
bytes, _ := json.Marshal(obj)
```

### 2. Validate early

```go
// Validation at .Set() time
action := sdto.NewAction("transfer", schema).
    Set("amount", -100)  // ❌ Fails immediately

// Errors collected at .Finalize()
saeBytes, err := action.Finalize()
```

### 3. Backend verifies, never repairs

```go
// ✅ Backend verifies
err := vax.VerifyChain(...)
err = sdto.ValidateData(...)

// ❌ Backend never "fixes" client data
// Don't do: data["amount"] = abs(data["amount"])
```

---

## Performance

Pure Go implementation with no dependencies:

```
BenchmarkComputeSAI-8           50000    24.3 µs/op
BenchmarkComputeGenesisSAI-8   100000    12.1 µs/op
BenchmarkJCSMarshal-8          200000     8.5 µs/op
```

---

## Cross-Language Compatibility

All implementations produce identical output:

```bash
# Test vector
actorID: "user123:device456"
genesisSalt: a1a2a3a4a5a6a7a8a9aaabacadaeafb0

# Expected
genesisSAI: afc50728cd79e805a8ae06875a1ddf78ca11b0d56ec300b160fb71f50ce658c3

# Verify
go test ./pkg/vax -v -run TestGenesisSAI
```

---

## Documentation

- [Architecture](../docs/ARCHITECTURE.md)
- [L0 Specification](../docs/SPECIFICATION.md)
- [Changelog](doct/changelog.md)

---

## License

MIT License

---

## Philosophy

> VAX is a tool, not a system.
>
> We provide primitives.
> You decide how to use them.
