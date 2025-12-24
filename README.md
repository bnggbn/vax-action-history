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

### Using C Toolkit
```bash
cd c
make
./vax-cli verify action.json
```

### Using Go Server
```bash
cd go
go run ./cmd/vaxd -config config.yaml
```

---

## Core Concepts

- **Action Object** — Canonical representation of a semantic action
- **Action Hash (SAI)** — Cryptographic proof: `H(prevSAI || SAE || gi)`
- **Actor Chain** — One `(user, device)` = one linear history
- **SO Factory** — Backend-defined semantic normalization

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

## Documentation

- 🏗️ [Architecture & Design Philosophy](docs/ARCHITECTURE.md)
- 📋 [L0 Specification](docs/SPECIFICATION.md)
- 🔧 [C API Reference](docs/C_API.md)
- 🚀 [Go API Reference](docs/GO_API.md)
- 🌐 [Deployment Guide](docs/DEPLOYMENT.md)

---

## Directory Layout

```
vax/
├── docs/          # Shared documentation
├── c/             # C toolkit (core primitives)
├── go/            # Go server (reference implementation)
├── examples/      # Integration examples
└── scripts/       # Build & test tools
```

See full structure in [Directory Layout](#directory-layout).

---

## Roadmap

### v0.6 (Current)
- [O] C core implementation
- [X] Go verifier server
- [ ] Comprehensive test vectors
- [ ] CLI tooling

### Future
- Cross-language SDK bindings
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