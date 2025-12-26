# VAX Go CGO Bindings

Go bindings for the VAX C core library.

## Prerequisites

1. **Build C library first**:
```bash
cd ../c
cmake -B build -G Ninja -DCMAKE_BUILD_TYPE=Release
cmake --build build
```

2. **Install OpenSSL** (if not already installed):
```bash
# Windows (vcpkg)
vcpkg install openssl:x64-windows

# Ubuntu/Debian
sudo apt-get install libssl-dev

# macOS
brew install openssl
```

## Usage

```go
package main

import (
    "fmt"
    "vax"
)

func main() {
    // Generate k_chain (session secret)
    kChain := make([]byte, vax.KChainSize)
    // ... fill with secure random bytes

    // Compute gi for counter=1
    gi, err := vax.ComputeGI(kChain, 1)
    if err != nil {
        panic(err)
    }
    fmt.Printf("gi: %x\n", gi)

    // Compute genesis SAI
    actorID := "user123:device456"
    genesisSalt := make([]byte, vax.GenesisSaltSize)
    // ... fill with random bytes

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
}
```

## Testing

```bash
# Run tests
go test -v

# Run with race detector
go test -race -v

# Run benchmarks
go test -bench=. -benchmem

# Test with C library coverage
go test -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Memory Safety

All CGO bindings properly manage memory:
- `C.CString()` results are freed with `C.free()`
- Fixed-size arrays (SAI, gi, etc.) are passed by reference
- No memory leaks (verified with AddressSanitizer)

## API

### Functions

#### `ComputeGI(kChain []byte, counter uint16) ([]byte, error)`
Compute gi_n = HMAC_SHA256(k_chain, "VAX-GI" || counter)

#### `ComputeSAI(prevSAI, saeBytes, gi []byte) ([]byte, error)`
Compute SAI_n = SHA256("VAX-SAI" || prevSAI || SHA256(SAE) || gi)

#### `ComputeGenesisSAI(actorID string, genesisSalt []byte) ([]byte, error)`
Compute genesis SAI_0 = SHA256("VAX-GENESIS" || actor_id || genesis_salt)

#### `VerifyAction(...) error`
Verify an action submission (crypto only, no JSON validation)

### Constants

- `SAISize = 32` - SAI length in bytes
- `GISize = 32` - gi length in bytes
- `KChainSize = 32` - k_chain length in bytes
- `GenesisSaltSize = 16` - genesis salt length in bytes

### Errors

- `ErrInvalidCounter` - Counter is not expected + 1
- `ErrInvalidPrevSAI` - prevSAI doesn't match expected
- `ErrSAIMismatch` - Computed SAI doesn't match submitted
- `ErrInvalidInput` - Invalid input parameters
- `ErrCounterOverflow` - Counter reached maximum value
- `ErrOutOfMemory` - Memory allocation failed

## Cross-Language Verification

All test vectors match the C test suite to ensure compatibility:

```bash
# Run C tests
cd ../c
ctest --test-dir build --output-on-failure

# Run Go tests
cd ../go
go test -v

# Both should produce identical outputs
```

## Performance

Typical benchmarks on modern hardware:
- `ComputeGI`: ~10 µs/op
- `ComputeSAI`: ~15 µs/op
- `VerifyAction`: ~25 µs/op

## Troubleshooting

### "cannot find -lvax"
Build the C library first:
```bash
cd ../c && cmake --build build
```

### "undefined reference to OpenSSL functions"
Install OpenSSL development libraries.

### CGO linking errors on Windows
Use vcpkg and specify the toolchain:
```bash
set CGO_CFLAGS=-I%VCPKG_ROOT%\installed\x64-windows\include
set CGO_LDFLAGS=-L%VCPKG_ROOT%\installed\x64-windows\lib
go test
```
