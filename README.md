# VAX — Semantic Action History (L0)

*A Git-like, tamper-evident history tool for actions, decisions, and transactions.*

---

## License

This project is licensed under the **MIT License**.  
You are free to use, modify, and distribute this software with attribution.

---

## What is VAX?

**VAX records actions the way Git records code.**

Instead of files and commits, VAX tracks **semantic actions** as immutable,
deterministic records and builds a verifiable history from them.

VAX is:
- Local-first
- Optional
- Incrementally adoptable

If you stop using VAX, your system still works.  
What remains is a **verifiable action history**.

---

## What VAX Provides

VAX focuses on **semantic integrity**, not enforcement.

It provides:

- Deterministic, **actor-bound action histories**
- Canonical JSON encoding
  - ASCII-only
  - No scientific notation
  - Deterministic key ordering
- Cryptographically verifiable records
- Cross-language reproducibility
- Back-end verification without trusting producers
- No dependency on blockchain or global consensus

VAX is designed for domains where **accountability matters**:
payments, financial flows, audits, risk decisions, and irreversible actions.

---

## Core Concepts

- **Action Object**  
  A canonical representation of a semantic action.

- **Action Hash**  
  Deterministically computed as:
H("VAX-SAI" || prev_hash || H(action_object) || gi)

- **Semantic Object (SO)**  
A strict schema defining allowed fields, types, ranges, and semantics.

- **IRP — Inverse Responsibility Principle**  
Semantics are defined by the backend; normalization happens at the producer.

- **Actor = One History**  
Each `(user_id, device_id)` pair produces a single linear action history.

---

## Distributed Usage Model

VAX is **not a distributed consensus system**.

Instead, VAX supports distributed environments by allowing
**multiple independent, verifiable action histories** to coexist.

Each Actor maintains a fully independent, linear history.
There is no global ordering, cross-Actor merging, or synchronization.

In distributed systems, nodes may:
- Maintain their own Actor histories
- Independently verify histories produced elsewhere
- Exchange verified action records without coordination

No central authority, shared state, or consensus mechanism is required.

This model avoids the complexity of distributed merging while still
providing strong guarantees of integrity, auditability, and determinism.

---

## Design Philosophy

VAX does not try to prevent mistakes.
It makes them **impossible to quietly rewrite**.

> *You may do the wrong thing — but you cannot pretend it never happened.*

Like Git, VAX does not enforce workflows.
It only guarantees history integrity.

---

## Roadmap (v0.3)

### Stage 1 — Core History Engine
- Action Object structure
- Canonical JSON builder (JCS-inspired)
- NFC normalization
- Illegal / shadow field rejection
- Action Hash generation with PRF-derived `gi`

### Stage 2 — Semantic Object Factory
- SO schema definition
- Type / min-max / enum validation
- Canonical payload construction
- Actor identity enforcement
- Explicit history switching

### Stage 3 — Testkit
- Canonical JSON determinism tests
- ASCII & scientific-notation rejection
- Action Hash test vectors
- History integrity tests (tamper / reorder / fork)
- Shadow-semantic detection

### Stage 4 — Tooling
- CLI utilities:
- `vax record`
- `vax verify`
- `vax audit`
- Debug & inspection modes
- Minimal runnable examples

### Stage 5 — Verification Backend (Optional)
- Semantic validation
- Canonical form verification
- History integrity checks
- Actor identity consistency
- Append-only history replica

---

## Directory Layout

vax/
├── cmd/
├── internal/
│ └── action/
├── test/
├── go.mod
├── go.sum
├── README.md


---

## Test Philosophy

VAX follows **byte-level determinism**:

> Given the same Semantic Object → Action Object → Action Hash pipeline,  
> outputs across languages must be identical.

If two independent implementations disagree, the implementation is wrong —
not the history.

---

## Scope & Non-Goals

VAX does **not** aim to:

- Replace real-time server-authoritative systems
- Enforce policy or block actions
- Act as a global standard or protocol

VAX exists to make **post-facto verification possible**
without trusting producers or operators.

---

For detailed construction rules and verification semantics,  
see `vax-l0-reference.md`.

---

## Contributing

Cross-language test vectors, semantic edge cases,
and audit tooling are welcome.

VAX grows by **usage**, not mandates.