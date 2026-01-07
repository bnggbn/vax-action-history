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
go get github.com/bnggbn/vax-action-history/go/pkg/vax
```

```go
package main

import (
    "fmt"
    "crypto/rand"
    "vax/pkg/vax"
)

func main() {
    // Compute genesis SAI
    actorID := "user123:device456"
    genesisSalt := make([]byte, 16)
    rand.Read(genesisSalt)

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

- **SAE** (Semantic Action Encoding) — Canonical JSON representation of an action
- **SAI** (Semantic Action Identifier) — Cryptographic hash: `SHA256("VAX-SAI" || prevSAI || SHA256(SAE))`
- **Actor Chain** — One `(user_id, device_id)` = one linear history
- **prevSAI** — Each action references its predecessor, forming an append-only chain

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

## How It Works

### 1. Genesis
```
SAI_0 = SHA256("VAX-GENESIS" || actor_id || genesis_salt)
```
Each Actor (user + device) starts with a unique genesis SAI.

### 2. Action Chain
```
SAI_n = SHA256("VAX-SAI" || prevSAI || SHA256(SAE))
```
Each subsequent action references the previous SAI, forming a tamper-evident chain.

### 3. Verification
Backend verifies:
- prevSAI continuity (no gaps or reordering)
- SAI computation correctness
- Schema compliance
- Backend signs the SAE to mark "action entered history"

---

## Implementation Status

| Language | Package | Status | Dependencies |
|----------|---------|--------|--------------|
| **Go** | `pkg/vax` | ✅ Complete | None (pure Go) |
| **C** | `libvax.a` | ✅ Complete | OpenSSL |
| **TypeScript** | `ts/` | ✅ Complete | None (pure TypeScript) |

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
- 🔧 [Go API Reference](go/README.md)
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
│   ├── pkg/vax/jcs/       # VAX-JCS canonicalizer
│   ├── pkg/vax/sae/       # SAE builder
│   └── pkg/vax/sdto/      # Schema-driven validation
└── ts/                # TypeScript implementation
    └── src/
        ├── jcs/           # JCS canonicalizer
        ├── sae/           # SAE builder
        ├── sdto/          # Schema-driven validation
        └── vax.ts         # Core primitives
```

---

## Running Tests

```bash
# Go
cd go && go test ./pkg/vax/...

# C
cd c && ctest --test-dir build

# TypeScript
cd ts && npm test
```

---

## Design Philosophy

### What VAX Provides
- **Append-only history**: Actions cannot be removed or reordered
- **Tamper-evident**: Any change to history is detectable
- **Local-first**: No coordination required between actors
- **Cross-language**: Deterministic results across implementations

### What VAX Does NOT Provide
- **Authorization**: VAX records what happened, not what's allowed
- **Conflict resolution**: Divergent histories are detected, not merged
- **Business logic**: Correctness is enforced at higher layers

### Defense in Depth
```
┌─────────────────────────┐
│   L2: Business Logic    │  ← Authorization, workflow
├─────────────────────────┤
│   L1: Semantic Layer    │  ← Schema, validation
├─────────────────────────┤
│   L0: VAX Integrity     │  ← Tamper evidence
├─────────────────────────┤
│   TLS                   │  ← Transport security
└─────────────────────────┘
```

---

## Roadmap

### v0.7 (Current)
- [x] C core implementation
- [x] Go pure implementation
- [x] TypeScript complete implementation
- [x] Cross-language test vectors
- [x] Schema-driven validation (SDTO)
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
