VAX L0 SDK — Reference Pipeline (Draft)

Purpose
Define the minimal, enforceable pipeline required to guarantee
“You cannot write history while standing on the past.”

This pipeline is structural, not a protocol and not a workflow.

# Phase 0 — Context Acquisition (Outside SDK)

Responsibility: Application / Infrastructure

The SDK does not read databases or coordinate states.

Before any action is constructed, the system MUST obtain:

baseSAI — latest committed SAI from the Real Replica

source may be DB / cache / API / replica

must reflect the current writable present

baseSAI := read_latest_sai_from_system()


This phase is mandatory but external to the SDK.

# Phase 1 — Intent Capture (Untrusted)

Responsibility: Application

Raw input may come from users, services, or automated systems

Input is considered semantically untrusted

No guarantees are assumed at this stage.

# Phase 2 — Semantic Normalization (SOF)

Responsibility: SDK / SO Factory

Validate input against backend-defined schema

Apply:

strict typing

range / enum checks

canonical field mapping

Reject:

unknown fields

shadow semantics

ambiguous representations

Output:

SDTO (Semantic Data Transfer Object)


Properties:

immutable

deterministic

free of business logic

safe to canonicalize

# Phase 3 — Canonical Encoding (SAE)

Responsibility: SDK

Transform SDTO into canonical byte representation:

ASCII-only

NFC normalized

deterministic key ordering

no scientific notation

no duplicate keys

Output:

SAE bytes


At this point, semantic ambiguity is destroyed.

# Phase 4 — Action Hash Preparation

Responsibility: SDK

Inputs:

baseSAI

SAE

locally derived metadata (counter, gi)

Steps:

gi := PRF(K_chain, counter)
SAI := H("VAX-SAI" || baseSAI || SAE || gi)


No state is mutated yet.

# Phase 5 — Write Gate Enforcement

Responsibility: SDK

Hard gate:

bool Commit(baseSAI)


Rules:

baseSAI MUST equal the latest committed SAI

SDK MUST NOT mutate application state before Commit succeeds

Results:

true → this action becomes the new writable present

false → conflict signal (no side effects)

This is the only point where history advances.

# Phase 6 — Observability / Emission

Responsibility: SDK

After a successful commit:

emit (SAI, SAE)

expose action metadata for:

logging

replication

auditing

downstream consumers

No guarantees required for delivery order.

# Phase 7 — State Mutation (Outside SDK)

Responsibility: Application

Only after Commit == true may the system:

mutate database state

trigger irreversible business effects

emit external side effects (payments, notifications, etc.)

SDK does not enforce this step —
the pipeline assumes it as a contract.

# Pipeline Invariants (Non-Negotiable)

1. History can only advance from the latest committed SAI

2. No action is valid without passing the write gate

3. Conflict is a signal, not a failure

4. Semantic equivalence ⇒ identical SAE bytes

5. SDK never repairs history

# Explicit Non-Goals (Pipeline Level)

- No consensus

- No leader election

- No merge or reconciliation

- No automatic retry strategy

- No distributed coordination

These belong to L1/L2.

# One-Sentence Summary

> VAX L0 destroys semantic ambiguity, enforces a single writable present,
> and emits immutable facts upon which higher-layer coordination may emerge.

---2025_11_29---