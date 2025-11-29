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

- **Semantic Object Factory** (SOF)
A backend-defined factory that validates, normalizes,
and produces immutable Semantic Data Transfer Objects (SDTO).

- **IRP — Inverse Responsibility Principle**  
Semantics are defined by the backend; normalization happens at the producer.

- **Actor = One History**  
Each `(user_id, device_id)` pair produces a single linear action history.

---

## Distributed Usage Model

VAX is **not a distributed consensus system**.

VAX does not attempt to establish global agreement,
leader election, quorum voting, or synchronized state.

Instead, VAX supports distributed environments by providing
**independent, verifiable action histories** that can be safely
produced, exchanged, and verified without coordination.

Each Actor maintains a strictly linear action history.
There is no global ordering across Actors, and no cross-Actor
merging or reconciliation.

In distributed systems, nodes may:

- Produce actions for their own Actors
- Verify action histories produced elsewhere
- Exchange or replicate verified actions asynchronously

VAX does not require a shared runtime state or
distributed coordination to perform verification.

This design avoids the complexity of consensus and distributed
merging, while still providing strong guarantees of:

- historical integrity  
- auditability  
- deterministic verification  

These properties allow VAX to operate **within** distributed
systems without becoming a distributed coordination mechanism.

VAX does not coordinate consensus.
It enforces a single writable present by design.
Consensus emerges naturally from this constraint.

VAX L0 is not an alternative to Raft.

It can be used **under** Raft (to make actions tamper-evident),
or **beside** Raft (to provide actor-level history integrity),
but it does not attempt to replace distributed log replication.

---

### Distributed Usage (SDK Behavior)

The VAX SDK is designed to behave correctly in distributed
deployments by enforcing a single invariant:

> **One Actor produces one canonical action history.**

Each SDK instance operates on exactly one Actor and produces
actions against a locally maintained execution context.

The SDK ensures that:

- each action references its immediate predecessor (`prevSAI`)
- action validity can be determined locally
- verification does not require global state or synchronization

In practice, this makes it possible for:

- multiple nodes to concurrently produce actions for different Actors
- verified actions to be transferred or replicated across systems
- receivers to verify actions out-of-band or asynchronously

The SDK does **not**:

- merge histories
- arbitrate conflicts
- compare actions across Actors

Conflict resolution and interpretation are explicitly delegated
to higher layers (L1/L2).

From an implementation perspective, an action emitted by the SDK
is verifiable given only its `SAE`, `SAI`, and `prevSAI`.

---

### Implementation Model

A minimal distributed verifier may be implemented using an
append-only key-value structure:


```

Map<SAI, SAE>

```

Verification proceeds as follows:

1. Extract `prevSAI` from the incoming action
2. Lookup `prevSAI` in local storage
3. Verify:
   - canonicalized SAE bytes
   - derived `gi`
   - computed `SAI`
4. Store `(SAI, SAE)` if verification succeeds

This enables O(1) local verification and allows
partial history replication without reprocessing
entire chains.

---

### Write Discipline

VAX L0 defines a strict write gate for establishing action history.

Before mutating persistent application state, systems MUST:

1. Submit the corresponding action to VAX L0
2. Receive a successful commit acknowledgment
3. Apply the derived state change

This constraint ensures that all persistent mutations
are backed by a canonical, verifiable action record.

Read paths, projections, and derived state rebuilding
remain unrestricted.


---

## Design Philosophy

VAX does not try to prevent mistakes.  
It makes them **impossible to quietly rewrite**.

> *You may do the wrong thing — but you cannot pretend it never happened.*

Like Git, VAX does not enforce workflows.
It only guarantees history integrity.

---

### Write Gate API

```
bool Commit(string baseSAI)
```
Contract:

baseSAI MUST be the latest committed SAI known to the caller

Returns true if the action is appended successfully

Returns false if baseSAI is stale or conflicts with history

The SDK MUST NOT mutate application state when Commit returns false

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