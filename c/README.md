# VAX C Core (libvax)

C reference implementation of VAX L0.

Minimal, deterministic, frozen. Defines **byte-level cryptographic facts**, not application logic.

## Scope

VAX L0 cryptographic primitives and verification:

- `gi` derivation (HMAC-SHA256)
- `genesis_sai` computation
- `sai` hash chaining
- `vax_verify_action()`
- Golden test vectors for cross-language consistency

## Non-Scope

Intentionally does NOT implement:

- SAE construction, JSON parsing, canonicalization (VAX-JCS)
- Schema definition or validation
- Business logic, state mutation, storage, network/IO

**SAE is opaque canonical bytes** — canonicalization handled at application layer (Go server, JS/TS frontend).

## Stability

- Frozen as reference implementation
- Public APIs stable from v1.0.0
- Breaking changes require major version bump

## Usage

- Reference for other language implementations
- Low-level verifier in security-sensitive environments
- Test oracle for cross-language consistency

Not intended as a full SDK.

## Build

See `BUILD.md`.