# VAX Go Implementation

Pure Go implementation of VAX cryptographic primitives.

## Features

- ✅ **Pure Go** — No CGo, no C compiler required
- ✅ **Zero Dependencies** — Only uses Go standard library
- ✅ **Cross-Platform** — Windows, Linux, macOS out of the box
- ✅ **Simple Install** — Just `go get`

## Installation

```bash
go get github.com/anthropics/vax-action-history/go/pkg/vax
```

## Usage

```go
package main

import (
    "crypto/rand"
    "fmt"
    "vax/pkg/vax"
)

func main() {
    // Generate k_chain (session secret)
    kChain := make([]byte, vax.KChainSize)
    rand.Read(kChain)

    // Compute gi for counter=1
    gi, err := vax.ComputeGI(kChain, 1)
    if err != nil {
        panic(err)
    }
    fmt.Printf("gi: %x\n", gi)

    // Compute genesis SAI
    actorID := "user123:device456"
    genesisSalt := make([]byte, vax.GenesisSaltSize)
    rand.Read(genesisSalt)

    genesisSAI, err := vax.ComputeGenesisSAI(actorID, genesisSalt)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Genesis SAI: %x\n", genesisSAI)

    // Compute action SAI
    sae := []byte(`{"action":"test"}`)
    sai, err := vax.ComputeSAI(genesisSAI, sae, gi)
    if err != nil {
        panic(err)
    }
    fmt.Printf("SAI: %x\n", sai)

    // Verify action
    err = vax.VerifyAction(
        kChain,
        0,            // expected counter (last committed)
        genesisSAI,   // expected prevSAI
        1,            // submitted counter
        genesisSAI,   // submitted prevSAI
        sae,          // SAE bytes
        sai,          // submitted SAI
    )
    if err != nil {
        fmt.Printf("Verification failed: %v\n", err)
    } else {
        fmt.Println("✓ Action verified!")
    }
}
```

## Testing

```bash
cd go

# Run all tests
go test ./...

# Run with verbose output
go test -v ./pkg/vax

# Run with race detector
go test -race ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## API Reference

### Functions

#### `ComputeGI(kChain []byte, counter uint16) ([]byte, error)`

Compute per-action entropy:
```
gi_n = HMAC_SHA256(k_chain, "VAX-GI" || counter)
```

**Parameters:**
- `kChain`: 32-byte session secret
- `counter`: Action counter (big-endian uint16)

**Returns:** 32-byte gi value

---

#### `ComputeSAI(prevSAI, saeBytes, gi []byte) ([]byte, error)`

Compute action hash using two-stage hash:
```
sae_hash = SHA256(SAE)
SAI_n = SHA256("VAX-SAI" || prevSAI || sae_hash || gi)
```

**Parameters:**
- `prevSAI`: Previous action's SAI (32 bytes)
- `saeBytes`: Canonical JSON bytes (VAX-JCS)
- `gi`: Per-action entropy (32 bytes)

**Returns:** 32-byte SAI value

---

#### `ComputeGenesisSAI(actorID string, genesisSalt []byte) ([]byte, error)`

Compute genesis SAI for a new actor chain:
```
SAI_0 = SHA256("VAX-GENESIS" || actor_id || genesis_salt)
```

**Parameters:**
- `actorID`: Actor identifier (e.g., "user123:device456")
- `genesisSalt`: 16 random bytes (must be persisted)

**Returns:** 32-byte genesis SAI

---

#### `VerifyAction(...) error`

Verify an action submission (crypto only, no JSON validation).

```go
func VerifyAction(
    kChain []byte,           // Session secret
    expectedCounter uint16,  // Last committed counter
    expectedPrevSAI []byte,  // Last committed SAI
    counter uint16,          // Submitted counter
    prevSAI []byte,          // Submitted prevSAI
    saeBytes []byte,         // SAE bytes
    sai []byte,              // Submitted SAI
) error
```

**Verification steps:**
1. Check counter overflow
2. Verify counter is `expected + 1`
3. Verify prevSAI matches
4. Recompute gi and SAI
5. Verify SAI matches

---

### Constants

```go
const (
    SAISize         = 32  // SAI length in bytes
    GISize          = 32  // gi length in bytes
    KChainSize      = 32  // k_chain length in bytes
    GenesisSaltSize = 16  // genesis salt length in bytes
)
```

### Errors

| Error | Description |
|-------|-------------|
| `ErrInvalidCounter` | Counter is not `expected + 1` |
| `ErrInvalidPrevSAI` | prevSAI doesn't match expected |
| `ErrSAIMismatch` | Computed SAI doesn't match submitted |
| `ErrInvalidInput` | Invalid input parameters (wrong length) |
| `ErrCounterOverflow` | Counter reached maximum value (65535) |

## Cross-Language Verification

Go implementation produces identical outputs to the C reference:

```bash
# Run C tests
cd c
ctest --test-dir build --output-on-failure

# Run Go tests
cd go
go test -v ./pkg/vax

# Known test vector (genesis SAI)
# actor_id: "user123:device456"
# genesis_salt: a1a2a3a4a5a6a7a8a9aaabacadaeafb0
# Expected SAI: afc50728cd79e805a8ae06875a1ddf78ca11b0d56ec300b160fb71f50ce658c3
```

## Performance

Benchmarks on modern hardware (Apple M1):

| Function | Time | Allocations |
|----------|------|-------------|
| `ComputeGI` | ~500 ns/op | 2 allocs |
| `ComputeSAI` | ~800 ns/op | 3 allocs |
| `ComputeGenesisSAI` | ~600 ns/op | 2 allocs |
| `VerifyAction` | ~1.5 µs/op | 5 allocs |

## Why Pure Go?

The original plan was to use CGo to bind the C library. However:

| Aspect | CGo | Pure Go |
|--------|-----|---------|
| **Dependencies** | C compiler + OpenSSL | None |
| **Cross-compile** | ❌ Complex | ✅ Trivial |
| **Build time** | Slow | Fast |
| **Deployment** | Complex | Single binary |
| **User experience** | `apt install gcc libssl-dev` | `go get` |

Go's `crypto/sha256` and `crypto/hmac` are highly optimized and provide
identical security guarantees to OpenSSL.

## Related Packages

- `vax/internal/jcs` — VAX-JCS canonical JSON encoder
- `vax/internal/sae` — Semantic Action Envelope builder
