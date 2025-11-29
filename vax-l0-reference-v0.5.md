# VAX L0 — Semantic Action History
**Reference Specification & Design Notes**

Version: 0.4  
Date: 2025-01-02  
Status: Draft (Non-Normative Reference)  
Author: bnggbn

---

## ⚠️ Status & Scope

This document is a **reference specification**, not a mandatory protocol.

It describes:
- canonical formats
- deterministic construction rules
- verification semantics

It does **not** require global adoption, negotiation, or standard compliance.

Implementations may adopt VAX L0 **partially or locally**.

---

## Table of Contents

1. Introduction  
2. Terminology  
3. Threat Model  
4. Distributed Context Assumptions 
5. Architectural Overview  
6. Core Data Model  
7. Canonicalization (VAX-JCS)  
8. Session Establishment  
9. Counter Semantics  
10. gi Derivation  
11. Action Hash Construction  
12. Submission Format  
13. State Machine  
14. Concurrency Model  
15. Schema System  
16. Security Properties  
17. Non-Goals  
18. Deployment Profiles (Informative)  
19. Security Considerations  
20. Test Vectors  
21. Key Words (RFC 2119)

---

## 1. Introduction

VAX L0 defines a **semantic integrity format** for recording actions
as **deterministic, append-only history**.

It provides:

- Deterministic action representation  
- Cross-language canonical encoding  
- Tamper-evident sequencing  
- Forward-secure append-only logs  

VAX L0 does **not** assume:
- trusted clients
- trusted frontends
- blockchain or consensus systems

Higher layers (L1/L2) remain responsible for:
authorization, business logic, and policy enforcement.

---

## 2. Terminology

The following terms are used in this document:

- **SO** — Semantic Object (schema-defined meaning)
- **SDTO** — Immutable semantic data representation
- **SAE** — Canonical encoded representation (VAX-JCS)
- **SAI** — Action Hash
- **K_chain** — Session secret (32 bytes)
- **gi_n** — PRF-derived per-action entropy
- **Actor** — `(user_id, device_id)`

The terms **MUST**, **SHOULD**, etc. follow RFC 2119,
but are used here to describe **reference-level behavior**.

---

## 3. Threat Model

### 3.1 Defended Against

- Replay and reordering
- Omission
- Hash-chain forgery
- Canonicalization bypass
- Shadow semantics
- Unicode drift
- Multi-node divergence
- Cross-language mismatch

### 3.2 Explicitly Out of Scope

- Client compromise
- Network MITM without TLS
- Malicious servers
- Authorization or business fraud

---

## 4. Distributed Context Assumptions

VAX L0 is designed around a single, deliberate constraint:
**one Actor produces one strictly linear action history**.

VAX does not attempt to establish:
- global ordering across Actors
- cross-Actor conflict resolution
- action merging or reconciliation
- distributed consensus

Instead, VAX provides a model where multiple independent action histories
can be verified and analyzed in a distributed environment without
coordination.

Each action references its immediate predecessor via `prevSAI`.
Verification is therefore **local and compositional**:
given `(prevSAI, SAE, gi)`, correctness can be determined without
global state.

As a result:
- Any node may independently verify an action history
- Histories may be stored, exchanged, or partially replicated
- Missing predecessors render actions *unverifiable*, not invalid

This design explicitly avoids the complexity of distributed merging
while preserving strong integrity and audit properties.

VAX is therefore **not a distributed protocol**, but a system whose
verification model remains valid under distribution.


---

## 5. Architectural Overview

Raw Input
↓
SO Factory (normalize + validate)
↓
SDTO (immutable)
↓
SAE (canonical encoding)
↓
Counter prediction (+1)
↓
gi = HMAC(K_chain, "VAX-GI" || counter)
↓
SAI = HASH(prevSAI || SAE || gi)
↓
(optional signature)
↓
Submit → Verify → Append


Backends **MUST NOT** modify semantics  
Backends **ONLY verify** and append

---

## 6. Core Data Model

### 6.1 Semantic Object (SO)

An SO is a schema-defined semantic unit.

Properties:
- No business logic
- Fully validated
- Canonicalizable
- Deterministic meaning

### 6.2 SDTO

An SDTO is an immutable output produced **only** by the SO Factory.

Producers MUST NOT manually construct SAE.

---

## 7. Canonicalization — VAX-JCS

This section defines **byte-level canonical rules**.

SAE encoding MUST satisfy:

- ASCII-only
- Deterministic key ordering
- No scientific notation
- NFC normalization
- UTF-16 escaping for non-ASCII

Rejected inputs:

- Duplicate keys
- NaN / Infinity
- Unpaired surrogates
- Invalid UTF-8

Equivalent semantics MUST produce identical bytes.

---

## 8. Session Establishment

Backend returns:

- session_id  
- k_chain (32 bytes)  
- counter_offset (uint16)  
- last_sai  
- genesis_salt  

### 8.1 Genesis Hash

SAI_0 = HASH("VAX-GENESIS" || actor_id || genesis_salt)

genesis_salt MUST be persisted.

---

## 9. Counter Semantics

### 9.1 Actor-Scoped Counter

Counter is **actor-scoped**, not session-scoped.

Session applies only a public offset.

### 9.2 Strict +1 Rule

expected = last + 1
if counter != expected → reject

### 9.3 SDK Behavior

- Predict counter
- Commit locally **only after ACK**
- Do not consume counter on failure

---

## 10. gi Derivation

gi_n = HMAC_SHA256(K_chain, "VAX-GI" || counter_n)

---

## 11. Action Hash Construction

SAI_n = SHA256(
"VAX-SAI" ||
prev_SAI ||
SAE ||
gi_n
)

Prefixes prevent ambiguity.

---

## 12. Submission Format (Reference)

Headers:
- X-VAX-Session
- X-VAX-Counter
- X-VAX-PrevSAI

Body:
```json
{
  "sae": "...",
  "sai": "...",
  "signature": "..." // optional
}
```
## 13. State Machine
CONNECTED → SYNCED → PROPOSING → COMMITTED
Failures result in reconnect & resync.

## 14. Concurrency Model
One Actor = One serialized history.

Backends should lock per-Actor during commit.

## 15. Schema System
Backend defines schema

Producer normalizes

Backend verifies only

Repairs are NOT permitted.

## 16. Security Properties
Provides:

Append-only history

Replay resistance

Canonical determinism

Does NOT provide:

Legal non-repudiation

Business correctness

## 17. Non-Goals
VAX L0 does not attempt to:

Define authorization

Enforce policies

Act as a global protocol

Replace L1/L2 checks

## 18. Deployment Profiles
Informative only.

Strict mode recommended.
Relaxed profiles MAY log anomalies.

## 19. Security Considerations
TLS required

K_chain must remain in memory

Counter overflow must be prevented

## 20. Test Vectors
Test vectors from v0.3 remain valid.

## 21. RFC 2119 Keywords
Used as reference semantics, not governance mandates.


