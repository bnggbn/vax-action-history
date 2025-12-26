# VAX — Verifiable Action History

**VAX records actions the way Git records code.**

Local-first, incrementally adoptable, no blockchain required.

---

## What is VAX?

VAX provides tools for building deterministic, tamper-evident action logs.

Like Git ensures code history integrity, VAX ensures action history integrity.

**Key properties:**
- Each action gets a deterministic hash
- Actions form an append-only chain
- History is verifiable without trusting producers
- No global consensus required

---

## Use Cases

VAX is designed for domains where **accountability matters**:
- Financial transactions
- Audit trails
- Risk decisions
- Irreversible actions

---

## Quick Start

### Go (Recommended)

```bash
go get github.com/anthropics/vax-action-history/go/pkg/vax
```

```go
package main

import (
    "fmt"
    "vax/pkg/vax"
)

func main() {
    // Compute genesis SAI
    actorID := "user123:device456"
    genesisSalt := []byte{0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8,
                          0xa9, 0xaa, 0xab, 0xac, 0xad, 0xae, 0xaf, 0xb0}

    sai, _ := vax.ComputeGenesisSAI(actorID, genesisSalt)
    fmt.Printf("Genesis SAI: %x\n", sai)
}
```

### TypeScript

```bash
npm install vax
```

```typescript
import { canonicalize } from 'vax/jcs';

const sae = canonicalize({ action: "test", amount: 100 });
console.log(sae); // {"action":"test","amount":100}
```

### C (Reference Implementation)

```bash
cd c
cmake -B build -G Ninja
cmake --build build
```

See [C Build Instructions](c/BUILD.md) for details.

---

## Core Concepts

- **SAE** (Semantic Action Encoding) — Canonical JSON representation
- **SAI** (Semantic Action Identifier) — Cryptographic hash: `SHA256("VAX-SAI" || prevSAI || SHA256(SAE) || gi)`
- **Actor Chain** — One `(user, device)` = one linear history
- **gi** — Per-action entropy: `HMAC(k_chain, "VAX-GI" || counter)`

---

## Architecture

VAX is **not a protocol**, it's a **tool**.

Like Git:
- Git doesn't enforce workflows → VAX doesn't enforce policies
- Git guarantees history integrity → VAX guarantees action integrity
- Git is local-first → VAX is local-first

**Key principle:**
> You may do the wrong thing — but you cannot pretend it never happened.

See [Architecture & Philosophy](docs/ARCHITECTURE.md) for design rationale.

---

## Implementation Status

| Language | Package | Status | Dependencies |
|----------|---------|--------|--------------|
| **Go** | `pkg/vax` | ✅ Complete | None (pure Go) |
| **C** | `libvax.a` | ✅ Complete | OpenSSL |
| **TypeScript** | `ts/` | ✅ JCS Complete | None |

### Cross-Language Verification

All implementations produce identical outputs:

```
# Genesis SAI test vector
actor_id: "user123:device456"
genesis_salt: a1a2a3a4a5a6a7a8a9aaabacadaeafb0
Expected: afc50728cd79e805a8ae06875a1ddf78ca11b0d56ec300b160fb71f50ce658c3
```

---

## Documentation

- 🏗️ [Architecture & Design Philosophy](docs/ARCHITECTURE.md)
- 📋 [L0 Technical Specification](docs/SPECIFICATION.md)
- 🔧 [Go API Reference](go/cmd/doct/VAX_GO.md)
- 🔨 [C Build Instructions](c/BUILD.md)

---

## Directory Layout

```
vax/
├── docs/              # Shared documentation
│   ├── ARCHITECTURE.md    # Design philosophy
│   └── SPECIFICATION.md   # L0 technical spec
├── c/                 # C reference implementation
│   ├── include/vax.h      # Public API
│   ├── src/               # Implementation
│   └── test/              # Test suite
├── go/                # Go implementation (pure Go)
│   ├── pkg/vax/           # Core cryptographic primitives
│   ├── internal/jcs/      # VAX-JCS canonicalizer
│   └── internal/sae/      # SAE builder
└── ts/                # TypeScript implementation
    └── src/               # JCS canonicalizer
```

---

## Running Tests

```bash
# Go
cd go && go test ./...

# C
cd c && ctest --test-dir build

# TypeScript
cd ts && npm test
```

---

## Roadmap

### v0.6 (Current)
- [x] C core implementation
- [x] Go pure implementation
- [x] TypeScript JCS
- [x] Cross-language test vectors
- [ ] CLI tooling

### Future
- Python bindings
- Audit visualization tools
- Performance benchmarks

---

## License

MIT License — Free to use, modify, and distribute with attribution.

---

## Contributing

Cross-language test vectors, semantic edge cases,
and audit tooling are welcome.

VAX grows by **usage**, not mandates.
