# VAX — Verifiable Action History

**VAX records actions the way Git records code.**

A minimal toolkit for tamper-evident action logs. Local-first, incrementally adoptable.

---

## What is VAX?

VAX provides **cryptographic primitives** for building deterministic, tamper-evident action logs.

Like Git ensures code history integrity, VAX ensures action history integrity.

**Core primitives:**
- **SAI Chain**: Merkle-like hash chain for tamper detection
- **JCS**: Canonical JSON encoding for deterministic output
- **SDTO**: Schema-driven validation for semantic correctness

**What VAX is NOT:**
- ❌ Not a complete system
- ❌ Not a database
- ❌ Not a network protocol
- ❌ Not an authorization framework

VAX is a **tool**. You decide how to use it.

---

## Philosophy: Tool, Not System

```
Git doesn't force:
  - Where to store repos
  - How to sign commits
  - Which workflow to use

VAX doesn't force:
  - Where to store actions
  - Whether to add signatures
  - Which transport protocol to use
```

**You control:**
- Storage structure
- Signature mechanism (if needed)
- Authorization logic
- Business rules

**VAX provides:**
- Tamper-evident hash chain
- Deterministic encoding
- Schema validation

---

## Use Cases

VAX is designed for domains where **accountability matters**:

- Financial transactions
- Audit trails
- Risk decisions
- Irreversible actions
- Multi-party workflows

---

## Quick Start

### Installation

```bash
# Go
go get github.com/bnggbn/vax-action-history/go/pkg/vax

# TypeScript
npm install vax

# C
See c/BUILD.md
```

### Basic Usage

```go
import "vax/pkg/vax"

// 1. Compute genesis SAI
actorID := "user123:device456"
genesisSalt := vax.GenerateGenesisSalt()
genesisSAI, _ := vax.ComputeGenesisSAI(actorID, genesisSalt)

// 2. Build action
saeBytes := buildSAE("transfer", data)

// 3. Compute SAI
sai, _ := vax.ComputeSAI(prevSAI, saeBytes)

// 4. Verify chain (backend)
err := vax.VerifyChain(expectedPrevSAI, saeBytes, clientSAI)
if err != nil {
    // Chain broken or SAI mismatch
}

// 5. Store (you decide the structure)
db.Store(ActionRecord{
    SAI:     sai,
    SAE:     saeBytes,
    PrevSAI: prevSAI,
    // Add whatever else you need
})
```

---

## Core Concepts

### SAI Chain (Merkle-like)

```
SAI_0 = SHA256("VAX-GENESIS" || actorID || genesisSalt)
SAI_n = SHA256("VAX-SAI" || prevSAI || SHA256(SAE))
```

Each action references its predecessor, forming an append-only chain.

**Properties:**
- Tamper-evident: Modifying any action breaks the chain
- Deterministic: Same input → same SAI
- Local: No coordination required

### SAE (Semantic Action Encoding)

Canonical JSON representation of an action:

```json
{
  "action_type": "transfer",
  "timestamp": 1704672000000,
  "sdto": {
    "username": "alice",
    "amount": 500,
    "currency": "USD"
  }
}
```

**Always use JCS (JSON Canonicalization Scheme):**
- UTF-8 encoding
- Lexicographic key ordering
- No whitespace
- Deterministic number format

### Actor

An `(user_id, device_id)` pair with one linear history.

```
actorID = "user123:device456"
```

Each Actor maintains:
- One genesis SAI
- One append-only chain

---

## Optional: Add Signatures

VAX doesn't handle signatures. Use standard libraries:

```go
// Sign SAE (recommended)
signature := ed25519.Sign(privateKey, saeBytes)

// Or sign SAI
signature := ed25519.Sign(privateKey, sai)

// Store with signature
db.Store(ActionRecord{
    SAI:       sai,
    SAE:       saeBytes,
    PrevSAI:   prevSAI,
    Signature: signature,  // Your choice
})
```

**When to add signatures:**
- Legal non-repudiation required
- Multi-party accountability
- External audit requirements

**When NOT to add signatures:**
- Internal audit logs
- Trusted environment
- Performance-critical scenarios

---

## Architecture

VAX operates at Layer 0 (L0) - integrity only.

```
┌─────────────────────────┐
│   L2: Business Logic    │  ← Authorization, workflow
├─────────────────────────┤
│   L1: Semantic Layer    │  ← Schema, validation
├─────────────────────────┤
│   L0: VAX Integrity     │  ← Tamper evidence (SAI chain)
├─────────────────────────┤
│   TLS                   │  ← Transport security
└─────────────────────────┘
```

**VAX Guarantees:**
- ✅ Tamper detection via SAI chain
- ✅ Deterministic encoding via JCS
- ✅ Schema compliance via SDTO

**VAX Does NOT Guarantee:**
- ❌ Authorization (use L1/L2)
- ❌ Signatures (add if needed)
- ❌ Storage durability (use your DB)
- ❌ Network security (use TLS)

---

## Implementation Status

| Language | Package | Status | Dependencies |
|----------|---------|--------|--------------|
| **Go** | `pkg/vax` | ✅ Complete | None (pure Go) |
| **C** | `libvax.a` | ✅ Complete | OpenSSL |
| **TypeScript** | `ts/` | ✅ Complete | None (pure TS) |

### Cross-Language Verification

All implementations produce identical outputs:

```
# Genesis SAI test vector
actorID: "user123:device456"
genesisSalt: a1a2a3a4a5a6a7a8a9aaabacadaeafb0
Expected SAI: afc50728cd79e805a8ae06875a1ddf78ca11b0d56ec300b160fb71f50ce658c3
```

---

## Examples

See [examples/](examples/) for common patterns:

- **[no-signature](examples/no-signature/)** - Minimal setup, no signatures
- **[backend-signature](examples/backend-signature/)** - Backend signs SAE
- **[client-signature](examples/client-signature/)** - Client signs SAE
- **[multi-party](examples/multi-party/)** - Multiple signatures per action
- **[replica-verification](examples/replica-verification/)** - Cross-org verification

---

## Documentation

- 🏗️ [Architecture & Design Philosophy](docs/ARCHITECTURE.md)
- 📋 [L0 Specification](docs/SPECIFICATION.md)
- 🔧 [Go SDK](go/README.md)
- 🔧 [TypeScript SDK](ts/README.md)
- 🔨 [C Build Instructions](c/BUILD.md)

---

## Design Principles

### 1. Minimal

Only cryptographic primitives. No opinions on storage, transport, or authorization.

### 2. Deterministic

Same input always produces same output. Verifiable across implementations and time.

### 3. Local-First

No coordination required. Each Actor maintains independent history.

### 4. Incrementally Adoptable

Start simple, add complexity only when needed.

---

## Comparison

### VAX vs Git

| Aspect | VAX | Git |
|--------|-----|-----|
| **Target** | Actions | Files |
| **Hash Chain** | SAI chain | Commit chain |
| **Signatures** | Optional | Optional (GPG) |
| **Branching** | No | Yes |
| **Philosophy** | Tool, not protocol | Tool, not protocol |

### VAX vs Blockchain

| Aspect | VAX | Blockchain |
|--------|-----|------------|
| **Consensus** | None | Global (PoW/PoS) |
| **Scalability** | O(1) per Actor | O(network) |
| **Finality** | Immediate | Delayed |
| **Adoption** | Incremental | All-or-nothing |

### VAX vs Event Sourcing

| Aspect | VAX | Event Sourcing |
|--------|-----|----------------|
| **Focus** | Integrity | State reconstruction |
| **Verification** | Cryptographic | Replay-based |
| **Determinism** | Byte-level | Semantic-level |

---

## Roadmap

### v0.8 (Current)
- [x] Simplified API (removed signatures)
- [x] Pure tool philosophy
- [x] Cross-language implementations
- [x] Complete documentation

### Future
- [ ] CLI tooling
- [ ] Audit visualization
- [ ] Performance benchmarks
- [ ] Python bindings

---

## Testing

```bash
# Go
cd go && go test ./pkg/vax/...

# C
cd c && ctest --test-dir build

# TypeScript
cd ts && npm test
```

---

## Contributing

VAX grows by **usage**, not mandates.

Contributions welcome:
- Cross-language test vectors
- Real-world use cases
- Performance improvements
- Additional language bindings

---

## License

MIT License — Free to use, modify, and distribute with attribution.

---

## Credits

Inspired by:
- Git's design philosophy
- Merkle trees
- CQRS/Event Sourcing
- DDD tactical patterns

**VAX = Verifiable Action eXchange**

Not a protocol. Not a system. Just a tool.

## Optional Go storage companion

[storage/](storage/README.md) is an independently versioned, standard-library-only content-addressed storage SDK. It provides UTF-8 text trees, opaque blobs, bounded catalogs and byte verification. The VAX core remains independent of storage, authorization and transport. Install with `go get github.com/bnggbn/vax-action-history/storage@v0.1.0`; storage tags use `storage/vX.Y.Z` and do not publish C or npm packages. PFCO/PFCA formats remain compatible with Pageforge. This initial companion is Go-only.
