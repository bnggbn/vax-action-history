# VAX Architecture & Design Philosophy

**Version:** 0.7
**Date:** 2026-01-07
**Status:** Reference Design Document

---

## Table of Contents

1. Core Principle: Tool, Not Protocol
2. Why Not a Distributed Consensus System
3. Design Focus: Write Pipeline, Not Coordination
4. The IRP Principle (Inverse Responsibility)
5. Conflict and Divergence Philosophy
6. Security Properties and Their Limits
7. Comparison with Other Systems
8. Design Trade-offs
9. Threat Model Rationale
10. Non-Goals (and Why)

---

## 1. Core Principle: Tool, Not Protocol

VAX is a **reference implementation**, not a normative standard.

Like Git:
- Git is a tool for version control
- VAX is a tool for action history
- Neither defines "correct" usage
- Both guarantee integrity, not policy

**Key Insight:**

VAX does not prevent mistakes.
It makes them **impossible to quietly rewrite**.

> *You may do the wrong thing — but you cannot pretend it never happened.*

This is fundamentally different from systems that attempt to:
- prevent bad actions (authorization systems)
- define correct workflows (business logic)
- establish global truth (consensus protocols)

VAX only ensures that **what happened is verifiable**.

---

## 2. Why Not a Distributed Consensus System

VAX L0 is designed around a single, deliberate constraint:

> **One Actor produces one strictly linear action history.**

### What VAX Does NOT Attempt

VAX does not establish:
- Global ordering across Actors
- Cross-Actor conflict resolution
- Action merging or reconciliation
- Distributed consensus

### Why This Design?

**Reason 1: Complexity Avoidance**

Distributed consensus (Raft, Paxos, blockchain) is necessary when:
- Multiple writers need to agree on a single state
- No single authority can make decisions
- Network partitions must be tolerated

VAX sidesteps this by design:
- Each Actor is its own authority
- No cross-Actor coordination is needed
- Conflicts are **detected**, not resolved

**Reason 2: Verification Remains Local**

Each action references its immediate predecessor via `prevSAI`.

Verification is therefore local and compositional:
given `prevSAI` and the server’s current commit state, continuity can be determined without global coordination.

This means:
- Any node may independently verify chain continuity
- Histories may be stored, exchanged, or partially replicated
- Missing predecessors render actions *unverifiable*, not *invalid*

**Reason 3: Consensus Emerges Naturally**

VAX enforces a **single writable present** by design.

When multiple systems need agreement, they can use:
- Raft/Paxos for log replication
- Leader election for write authority
- Traditional locking for critical sections

VAX sits **under** these systems, making their outputs tamper-evident.

### Distributed Usage Model

In distributed environments, VAX SDK instances MAY exist on multiple nodes.

Each instance:
- Produces actions for exactly one Actor
- Maintains local chain state (prevSAI)
- Emits actions referencing `prevSAI`

The SDK makes no assumptions about:
- Message ordering
- Delivery guarantees
- Transport topology
- Centralized coordination

An action whose `prevSAI` is unknown is treated as **unverifiable**, not invalid.

This allows action histories to be verified incrementally in distributed or
partially replicated systems without chain merging.

---

## 3. Design Focus: Write Pipeline, Not Coordination

VAX L0 is an **action write pipeline**, not a coordination mechanism.

### Single Responsibility

VAX L0's sole responsibility:

> Decide whether an action is allowed to become part of persistent history.

VAX L0 does **not**:
- Resolve conflicts
- Repair state
- Merge histories
- Interpret semantics

It only **admits or rejects facts**.

### The Write Flow
```
Raw Input
    ↓
SO Factory (normalize + validate)
    ↓
SDTO (immutable)
    ↓
SAE (canonical encoding)
    ↓
SAI = SHA256("VAX-SAI" || prevSAI || SHA256(SAE))
    ↓
Action Envelope (SAE + SAI + prevSAI)
    ↓
Backend: VERIFY prevSAI continuity (never modify)
    ↓
Backend: VERIFY SAI computation
    ↓
Backend: VERIFY SDTO against Schema
    ↓
Backend: SIGN SAE (action enters history)
    ↓
COMMIT or REJECT
```

**Critical Constraint:**

Backends **MUST NOT** modify semantics.
Backends **ONLY verify** and append.

### Why This Constraint?

If backends could "fix" data:
- Producers lose determinism
- Cross-language verification breaks
- History becomes backend-dependent
- Audit trails become unreliable

Instead, VAX enforces:
- Producer normalizes → Backend verifies
- If normalization fails → Action is rejected
- No silent corrections → No hidden mutations

---

## 4. The IRP Principle (Inverse Responsibility)

**IRP = Inverse Responsibility Principle**

Traditional systems:
- Frontend sends raw data
- Backend cleans, validates, and normalizes
- Backend defines "correct" representation

VAX inverts this:
- Backend defines semantic schema
- **Producer** (frontend/SDK) normalizes to schema
- Backend verifies compliance, never repairs

### Why IRP?

**Problem with Traditional Model:**

If backends normalize data, then:
- Different backend versions may normalize differently
- Cross-language implementations diverge
- History becomes non-deterministic

**IRP Solution:**

The Semantic Object Factory (SOF) runs **producer-side**.

SOF responsibilities:
- Enforce schema-defined semantic constraints
- Reject invalid or ambiguous input
- Normalize input into canonical form

Backend responsibilities:
- Verify that normalization was done correctly
- Reject non-compliant actions
- **Never repair or modify**

### Example: Unicode Normalization

Traditional:
```
Frontend: sends "café" (NFC)
Backend: silently converts to NFD
Database: stores NFD
```

VAX IRP:
```
Frontend: MUST normalize to NFC before submission
Backend: verifies NFC compliance
If not NFC → reject with error
Database: stores exactly what frontend sent
```

This ensures:
- Go implementation produces identical bytes to C implementation
- History is reproducible across time and systems
- No silent mutations

---

## 5. Conflict and Divergence Philosophy

### Core Principle

> Divergence is not an error at L0.
> It is an observable fact.

VAX L0 deliberately avoids resolving conflicts or selecting a "correct" history.

The presence of multiple incompatible histories represents **what actually happened**,
not a protocol failure.

### What This Means

**A missing `prevSAI`:**
- Represents insufficient information
- NOT invalid input
- Action is **unverifiable**, not wrong

**Conflicting histories:**
- Indicate concurrent or inconsistent writes
- Are explicitly detectable
- Require higher-layer interpretation

**No automatic resolution:**
- L0 does not merge
- L0 does not choose
- L0 only verifies internal consistency

### Why Not Resolve Conflicts?

**Reason 1: Single Responsibility**

Conflict resolution requires:
- Business logic (which Actor wins?)
- Policy decisions (reject both? merge?)
- Authorization (who can override?)

These are **not** L0 concerns.

**Reason 2: Explicit is Better**

Automatic conflict resolution can:
- Hide problems
- Make wrong decisions
- Create silent data loss

Explicit detection allows:
- Human review
- Policy-based resolution
- Audit of resolution decisions

**Reason 3: Verification Remains Simple**

If L0 resolved conflicts, verification would require:
- Knowledge of resolution rules
- Access to all conflicting branches
- Complex state machines

By keeping L0 simple, verification remains:
- Local
- Deterministic
- Easy to implement

### L0 Guarantees

VAX L0 guarantees:
- Divergence **cannot occur silently**
- Incompatible histories are **explicitly detectable**
- All accepted history remains **internally consistent and verifiable**

How conflicts are interpreted, prioritized, or resolved is explicitly
**out of scope** for L0.

---

## 6. Security Properties and Their Limits

### What VAX Provides

**1. Append-Only History**
- Once committed, actions cannot be removed
- Tampering is detectable
- History is auditable

**2. Replay Resistance**
- Replays fail due to `prevSAI` continuity check
- Each action must reference the exact previous SAI
- Reordering or replaying breaks the chain


**3. Canonical Determinism**
- Same semantic input → same bytes
- Cross-language reproducibility
- No ambiguity in representation

**4. Tamper Evidence**
- Hash chain (`prevSAI` linking)
- Any modification breaks verification
- Missing actions are detectable (gap in chain)

### What VAX Does NOT Provide

**1. Legal Non-Repudiation**

VAX L0 uses hash chains for integrity, not cryptographic signatures.

Backend signatures (Ed25519) are added at verification time to mark
"action entered history", but client signatures are optional and
schema-dependent.

Why this design?
- Simpler L0 implementation
- Faster verification
- Adequate for most use cases
- Signatures can be added when needed

If legal non-repudiation from clients is required:
- Schema can mandate client signatures via `sign` or `sign_multi` fields
- Backend always signs to establish authority
- This is L1 concern, not L0

**2. Business Correctness**

VAX ensures actions are **recorded correctly**.

It does NOT ensure actions are:
- Authorized
- Sensible
- Profitable
- Legal

Example:
```
Action: "transfer $1M to attacker"
VAX: ✅ Valid action, correctly recorded
Business Logic: ❌ Unauthorized transfer
```

**3. Authorization**

VAX records **what happened**, not **what should happen**.

Authorization must be enforced at:
- L1 (semantic validation)
- L2 (business logic)
- Application layer

**4. Protection Against Compromised Clients**

If an attacker controls:
- The device
- The SDK
- The client's chain state (prevSAI)

Then they can produce valid actions.

VAX's defense:
- Compromised actions are still **recorded**
- History remains **auditable**
- Tampering is **detectable**

But VAX cannot prevent a compromised client from producing bad actions.

### Threat Model Summary

| Threat | Defended? | How? |
|--------|-----------|------|
| Replay attack | ✅ Yes | prevSAI chain |
| Reordering | ✅ Yes | prevSAI chain |
| Omission | ✅ Yes | Gap detection |
| Tampering | ✅ Yes | Hash verification |
| Client compromise | ❌ No | Out of scope |
| Network MITM | ⚠️ Partial | Requires TLS |
| Malicious server | ❌ No | Out of scope |
| Authorization bypass | ❌ No | L1/L2 concern |

---

## 7. Comparison with Other Systems

### VAX vs. Blockchain

| Aspect | VAX | Blockchain |
|--------|-----|------------|
| **Consensus** | None (single Actor authority) | Global (PoW/PoS/BFT) |
| **Purpose** | Action integrity | Decentralized trust |
| **Scalability** | O(1) per Actor | O(network size) |
| **Finality** | Immediate | Probabilistic or delayed |
| **Adoption** | Incremental | All-or-nothing |

**When to use Blockchain instead:**
- Need trustless environment
- Multiple untrusted parties
- Decentralization is core requirement

**When to use VAX instead:**
- Single organization
- Trust exists but needs verification
- Performance matters
- Incremental adoption needed

### VAX vs. Event Sourcing

| Aspect | VAX | Event Sourcing |
|--------|-----|----------------|
| **Focus** | Action integrity | State reconstruction |
| **Verification** | Cryptographic | Replay-based |
| **Determinism** | Byte-level | Semantic-level |
| **Cross-language** | Guaranteed | Not guaranteed |

**When to use Event Sourcing instead:**
- State rebuild is primary goal
- Complex aggregates
- Domain events are rich

**When to use VAX instead:**
- Cross-system verification needed
- Audit trail is critical
- Determinism is required

### VAX vs. Git

| Aspect | VAX | Git |
|--------|-----|-----|
| **Target** | Actions | Files |
| **Branching** | No (single Actor history) | Yes (multiple branches) |
| **Merging** | No | Yes |
| **Consensus** | None | None |
| **Purpose** | Action integrity | Change integrity |

**Key Similarity:**
Both are **tools for integrity**, not coordination protocols.

---

## 8. Design Trade-offs

### Trade-off 1: No Branching

**We chose:** Single linear history per Actor
**We gave up:** Branching and merging

**Why?**
- Branching requires conflict resolution
- Conflict resolution requires policy
- Policy is L1/L2 concern
- Keeping L0 simple

**When this hurts:**
- Multi-device scenarios without coordination
- Offline-first applications

**Mitigation:**
- Use device_id as part of Actor
- Higher layer can merge Actor histories

### Trade-off 2: Backend-Only Signatures (Client Signatures Optional)

**We chose:** Backend always signs, client signatures schema-dependent
**We gave up:** Mandatory client signatures for all actions

**Why?**
- Simpler client implementation
- Flexibility: schema decides when client signature is needed
- Backend signature establishes authority ("action entered history")
- Adequate for most use cases

**When this hurts:**
- Legal disputes requiring client non-repudiation
- Multi-party contracts where all parties must sign

**Mitigation:**
- Schema can mandate client signatures via `sign` or `sign_multi` fields
- Backend always signs to establish authority
- External notarization for legal requirements

### Trade-off 3: No Global Ordering

**We chose:** Per-Actor ordering only
**We gave up:** Cross-Actor causality

**Why?**
- Global ordering requires consensus
- Consensus adds complexity
- Most use cases don't need it

**When this hurts:**
- Distributed transactions
- Causal dependencies across Actors

**Mitigation:**
- Application-level causality tracking
- External coordination layer

### Trade-off 4: No Additional Entropy in SAI

**We chose:** SAI depends only on prevSAI and SAE
**We gave up:** Per-action randomness (gi)

**Why?**
- Simpler implementation and verification
- prevSAI chain provides replay protection
- Deterministic: same inputs → same SAI
- Easier to debug and reason about

**When this hurts:**
- SAI becomes predictable if SAE is known
- No additional entropy layer

**Mitigation:**
- prevSAI chain prevents replay/reordering
- Backend signatures add authority
- TLS protects transport
- Sufficient for most use cases

---

## 9. Threat Model Rationale

### Why These Boundaries?

**Defended Against:**
- Replay and reordering → Common attack vectors
- Omission → Silent data loss
- Hash-chain forgery → Tampering
- Canonicalization bypass → Cross-language divergence
- Shadow semantics → Hidden meanings
- Unicode drift → Subtle mutations

**Explicitly Out of Scope:**
- Client compromise → Cannot defend at protocol level
- Network MITM without TLS → Standard practice to require TLS
- Malicious servers → Trust boundary assumption
- Authorization fraud → L1/L2 responsibility

### Defense in Depth

VAX L0 is **one layer** in a defense strategy:
```
┌─────────────────────────┐
│   L2: Business Logic    │  ← Authorization, workflow
├─────────────────────────┤
│   L1: Semantic Layer    │  ← Schema, validation
├─────────────────────────┤
│   L0: VAX Integrity     │  ← Tamper evidence
├─────────────────────────┤
│   TLS                   │  ← Transport security
├─────────────────────────┤
│   OS / Hardware         │  ← Platform security
└─────────────────────────┘
```

Each layer has a specific responsibility.

VAX does not attempt to replace other layers.

---

## 10. Non-Goals (and Why)

### Non-Goal 1: Define Authorization

**Why not?**
- Authorization is domain-specific
- Requires business logic
- Changes over time
- VAX records what happened, not what's allowed

**What to use instead:**
- OAuth/OIDC for authentication
- RBAC/ABAC for authorization
- Application-level policy engines

### Non-Goal 2: Enforce Policies

**Why not?**
- Policies vary by domain
- Policies change
- Enforcement requires runtime context

**What VAX does:**
- Records when policy was violated
- Makes violations auditable
- Cannot be hidden after the fact

### Non-Goal 3: Act as a Global Protocol

**Why not?**
- Forces standardization
- Prevents innovation
- Creates committee bureaucracy

**What VAX is:**
- Reference implementation
- Best practices guide
- Foundation for custom solutions

### Non-Goal 4: Replace L1/L2 Checks

**Why not?**
- L0 focuses on integrity only
- Business logic belongs higher
- Separation of concerns

**Example:**
```python
# L2: Business Logic
if balance < amount:
    return "insufficient funds"

# L1: Semantic Validation
if not valid_account(account_id):
    return "invalid account"

# L0: VAX Integrity
action = vax.append(transfer_action)
# VAX doesn't care if transfer is valid business-wise
# It only cares that the action is recorded correctly
```

---

## Conclusion

VAX is designed to be:
- Simple (single responsibility)
- Local (no global coordination)
- Deterministic (cross-language reproducibility)
- Auditable (tamper-evident history)

It achieves this by explicitly **not** attempting to:
- Resolve conflicts
- Enforce policies
- Coordinate consensus
- Replace existing systems

VAX is a **tool**, not a **protocol**.

Use it where **integrity matters** and **simplicity is valuable**.
