# VAX L0 Technical Specification

**Version:** 0.6  
**Date:** 2025-12-08  
**Status:** Reference Implementation Behavior  
**Implementation:** vax-c, vax-go

---

## ⚠️ Document Status

This document describes the **precise technical behavior** of the VAX L0
reference implementations (vax-c and vax-go).

It is **not** a normative protocol specification.

For design rationale and philosophy, see [ARCHITECTURE.md](ARCHITECTURE.md).

---

## Table of Contents

1. Terminology
2. Data Formats
3. Canonicalization Rules (VAX-JCS)
4. Session Establishment
5. Counter Semantics
6. gi Derivation
7. Action Hash Construction
8. Submission Format
9. Verification Algorithm
10. State Machine
11. Concurrency Model
12. Error Codes
13. Test Vectors
14. RFC 2119 Usage

---

## 1. Terminology

The following terms are used throughout this specification:

- **SO** — Semantic Object (schema-defined meaning)
- **SDTO** — Semantic Data Transfer Object (immutable)
- **SAE** — Semantic Action Encoding (canonical JSON)
- **SAI** — Semantic Action Identifier (action hash, 32 bytes)
- **K_chain** — Session secret (32 bytes, MUST be memory-only)
- **gi_n** — PRF-derived per-action entropy (32 bytes)
- **Actor** — `(user_id, device_id)` tuple
- **Counter** — Strictly incrementing uint16 per Actor
- **prevSAI** — Previous action's SAI (32 bytes)

---

## 2. Data Formats

### 2.1 SAE (Semantic Action Encoding)

SAE is the canonical JSON representation of an action.

**Format Requirements:**

- **Encoding:** UTF-8
- **Whitespace:** None (compact form)
- **Key ordering:** Lexicographic (byte-order)
- **Numbers:** No scientific notation (e.g., `1e10` is invalid)
- **Unicode:** NFC normalization required
- **Non-ASCII:** Escaped as UTF-16 code units (`\uXXXX`)

**Example:**
```json
{"action":"transfer","amount":1000,"from":"alice","to":"bob"}
```

### 2.2 SAI (Semantic Action Identifier)

SAI is a 32-byte hash computed as:

**Two-stage hash (Implementation v2):**
```
sae_hash = SHA256(SAE)
SAI_n = SHA256(
    "VAX-SAI" ||
    prevSAI ||
    sae_hash ||
    gi_n
)
```

**Design rationale:**
- Avoids malloc for variable-length SAE
- Provides domain separation via "VAX-SAI" label
- Fixed-length message for the outer hash improves security

**Representation:**
- Binary: 32 bytes
- Hex: 64 characters (lowercase)

**Genesis SAI:**
```
SAI_0 = SHA256(
    "VAX-GENESIS" ||
    actor_id ||
    genesis_salt
)
```

Where:
- `actor_id` is UTF-8 encoded string
- `genesis_salt` is 16 random bytes (MUST be persisted)

### 2.3 gi (PRF-derived entropy)
```
gi_n = HMAC_SHA256(K_chain, "VAX-GI" || counter_n)
```

**Inputs:**
- `K_chain`: 32-byte session secret
- `"VAX-GI"`: ASCII prefix (6 bytes)
- `counter_n`: uint16 in big-endian

**Output:** 32 bytes

**Purpose:**
- Forward secrecy
- Backend participation in hash
- Replay prevention

---

## 3. Canonicalization Rules (VAX-JCS)

VAX-JCS is a strict subset of JSON with deterministic encoding rules.

### 3.1 Character Set

**Allowed:**
- ASCII printable: `0x20` - `0x7E`
- Control escapes: `\b \f \n \r \t`
- Unicode escapes: `\uXXXX` (UTF-16 code units)

**Rejected:**
- Raw non-ASCII characters
- Non-standard escapes
- Literal control characters

### 3.2 Numbers

**Format:**
```
-?(0|[1-9][0-9]*)(\.[0-9]+)?
```

**Rules:**
- No leading zeros (except `0` itself)
- No trailing zeros in decimals
- No scientific notation
- No `+` sign
- No `Infinity` or `NaN`

**Examples:**

| Input | Valid? | Canonical Form |
|-------|--------|----------------|
| `42` | ✅ | `42` |
| `042` | ❌ | — |
| `1.5` | ✅ | `1.5` |
| `1.50` | ❌ | — |
| `1e10` | ❌ | — |
| `-0` | ✅ | `-0` |

### 3.3 Strings

**Rules:**
- UTF-8 input MUST be NFC normalized
- Non-ASCII MUST be escaped as `\uXXXX`
- Control characters MUST be escaped
- No raw `U+0000` - `U+001F`

**Escape Sequences:**

| Character | Escape |
|-----------|--------|
| `"` | `\"` |
| `\` | `\\` |
| `/` | `\/` (optional) |
| Backspace | `\b` |
| Form feed | `\f` |
| Newline | `\n` |
| Carriage return | `\r` |
| Tab | `\t` |
| Unicode | `\uXXXX` |

**Example:**
```
Input:  "Hello, 世界"
Output: "Hello, \u4e16\u754c"
```

### 3.4 Objects

**Rules:**
- Keys MUST be sorted lexicographically (byte-order)
- No duplicate keys
- No whitespace
- Compact form

**Example:**
```json
{"action":"transfer","amount":100,"from":"alice"}
```

NOT:
```json
{
  "action": "transfer",
  "amount": 100,
  "from": "alice"
}
```

### 3.5 Arrays

**Rules:**
- Preserve order
- No whitespace
- Compact form

**Example:**
```json
[1,2,3]
```

### 3.6 Rejection Rules

The following inputs MUST be rejected:

- Non-UTF-8 byte sequences
- Non-NFC normalized strings
- Duplicate object keys
- Numbers in scientific notation
- Numbers with leading zeros
- `NaN`, `Infinity`, `-Infinity`
- Unpaired UTF-16 surrogates
- Invalid escape sequences

---

## 4. Session Establishment

### 4.1 Protocol Flow
```
Client → Server: POST /vax/session/init
                  Body: { "actor_id": "user123:device456" }

Server → Client: 200 OK
                 Body: {
                   "session_id": "uuid-v7",
                   "k_chain": "base64(32 bytes)",
                   "counter_offset": 0,
                   "last_sai": "hex(32 bytes)",
                   "genesis_salt": "base64(16 bytes)"
                 }
```

### 4.2 Session Data

**session_id:**
- UUID v7 (recommended) or UUID v4
- Used for session tracking only
- Not included in hash computation

**k_chain:**
- 32 random bytes
- MUST be generated server-side
- MUST be transmitted over TLS
- MUST be stored in memory only (SDK)
- MUST NOT be persisted to disk

**counter_offset:**
- uint16 (0-65535)
- Public offset for counter display
- Does not affect hash computation

**last_sai:**
- SAI of the last committed action
- Used as `prevSAI` for next action
- If no history exists, use genesis SAI

**genesis_salt:**
- 16 random bytes
- MUST be persistent (database)
- Used to compute genesis SAI
- Unique per Actor

### 4.3 Genesis Computation
```
SAI_0 = SHA256("VAX-GENESIS" || actor_id || genesis_salt)
```

**actor_id format:** `"{user_id}:{device_id}"`

**Example:**
```
actor_id = "user123:device456"
genesis_salt = random(16 bytes)
SAI_0 = SHA256("VAX-GENESIS" + "user123:device456" + genesis_salt)
```

---

## 5. Counter Semantics

### 5.1 Actor-Scoped Counter

Counter is **Actor-scoped**, not session-scoped.

Each `(user_id, device_id)` pair maintains its own counter.

**Properties:**
- Type: uint16 (0-65535)
- Initialization: Starts at 0 (or after genesis)
- Increment: Strictly +1
- Persistence: Backend MUST persist counter per Actor

### 5.2 Strict +1 Rule
```
expected_counter = last_committed_counter + 1
if incoming_counter != expected_counter:
    reject("ERR_INVALID_COUNTER")
```

**No gaps allowed.**

### 5.3 SDK Behavior

SDK MUST:
1. Predict next counter locally
2. Include counter in action submission
3. Commit counter **only after server ACK**
4. **Not** advance counter on failure
5. Resync with server after reconnect

**Example Flow:**
```
SDK local state: counter = 5, prevSAI = abc123

1. Predict: next_counter = 6
2. Build action with counter = 6
3. Submit to server
4. [Wait for response]

If success:
    SDK: counter = 6, prevSAI = def456

If failure (network error):
    SDK: counter = 5 (unchanged)
    Retry with counter = 6

If failure (counter mismatch):
    SDK: resync with server
    Fetch latest counter + SAI
```

### 5.4 Counter Overflow

When counter reaches 65535:

**Option 1: Reject further actions**
```
if counter == 65535:
    reject("ERR_COUNTER_OVERFLOW")
```

**Option 2: Start new Actor**
```
// Backend creates new Actor with fresh genesis
new_actor_id = "user123:device456-v2"
counter = 0
```

Choice depends on deployment policy (out of L0 scope).

---

## 6. gi Derivation

### 6.1 Algorithm
```
gi_n = HMAC_SHA256(K_chain, "VAX-GI" || counter_n)
```

**Inputs:**
- `K_chain`: 32-byte session secret
- `"VAX-GI"`: ASCII string (6 bytes: `0x56 0x41 0x58 0x2D 0x47 0x49`)
- `counter_n`: uint16 in big-endian (2 bytes)

**Output:** 32 bytes

### 6.2 Implementation Pseudocode
```python
def derive_gi(k_chain: bytes, counter: int) -> bytes:
    prefix = b"VAX-GI"
    counter_bytes = counter.to_bytes(2, byteorder='big')
    message = prefix + counter_bytes
    gi = hmac.new(k_chain, message, hashlib.sha256).digest()
    return gi
```

### 6.3 Test Vector
```
K_chain (hex):
  0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef

counter: 42

Message (hex):
  56 41 58 2D 47 49  00 2A
  |V |A |X |- |G |I  |counter=42

gi (hex):
  [compute and verify]
```

---

## 7. Action Hash Construction

### 7.1 SAI Algorithm
```
SAI_n = SHA256(
    "VAX-SAI" ||
    prevSAI ||
    SAE ||
    gi_n
)
```

**Component Sizes:**
- `"VAX-SAI"`: 7 bytes (ASCII)
- `prevSAI`: 32 bytes (raw binary)
- `SAE`: variable (UTF-8 JSON)
- `gi_n`: 32 bytes (raw binary)

**Total input:** 71 + len(SAE) bytes

### 7.2 Implementation Pseudocode
```python
def compute_sai(prev_sai: bytes, sae: str, gi: bytes) -> bytes:
    hasher = hashlib.sha256()
    hasher.update(b"VAX-SAI")
    hasher.update(prev_sai)
    hasher.update(sae.encode('utf-8'))
    hasher.update(gi)
    return hasher.digest()
```

### 7.3 Test Vector
```
prevSAI (hex):
  0000000000000000000000000000000000000000000000000000000000000000

SAE:
  {"action":"test"}

gi (hex):
  ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff

Expected SAI (hex):
  [compute and verify]
```

---

## 8. Submission Format

### 8.1 HTTP Request
```
POST /vax/action/submit
Content-Type: application/json
X-VAX-Session: <session_id>
X-VAX-Counter: <counter>
X-VAX-PrevSAI: <prev_sai_hex>

{
  "sae": "<canonical_json>",
  "sai": "<sai_hex>",
  "signature": "<optional_base64>"
}
```

### 8.2 Header Semantics

**X-VAX-Session:**
- UUID identifying the current session
- Used for routing only
- Not included in hash

**X-VAX-Counter:**
- Decimal representation of uint16
- Example: `"42"`

**X-VAX-PrevSAI:**
- Hex-encoded (lowercase, 64 chars)
- Example: `"abc123..."`

### 8.3 Body Fields

**sae:**
- Canonical JSON string
- MUST match VAX-JCS rules
- MUST be the exact bytes used in hash

**sai:**
- Hex-encoded SAI (lowercase, 64 chars)
- Server MUST recompute and verify

**signature:**
- Optional
- Base64-encoded digital signature
- L1 concern (out of L0 scope)

### 8.4 Response Format

**Success (200 OK):**
```json
{
  "status": "committed",
  "sai": "<sai_hex>",
  "counter": 42
}
```

**Failure (400 Bad Request):**
```json
{
  "status": "rejected",
  "error": "ERR_INVALID_COUNTER",
  "detail": "Expected counter 42, got 41"
}
```

---

## 9. Verification Algorithm

### 9.1 Server-Side Verification Steps
```
1. Parse Headers
   - Extract session_id, counter, prevSAI

2. Parse Body
   - Extract sae, sai, (optional signature)

3. Load Actor State
   - Fetch K_chain from session store
   - Fetch expected_counter from database
   - Fetch last_sai from database

4. Verify Counter
   if counter != expected_counter + 1:
       reject("ERR_INVALID_COUNTER")

5. Verify prevSAI
   if prevSAI != last_sai:
       reject("ERR_INVALID_PREV_SAI")

6. Verify SAE Canonicalization
   if not is_valid_vax_jcs(sae):
       reject("ERR_INVALID_CANONICALIZATION")

7. Derive gi
   computed_gi = HMAC(K_chain, "VAX-GI" || counter)

8. Recompute SAI
   computed_sai = SHA256("VAX-SAI" || prevSAI || sae || computed_gi)

9. Verify SAI Match
   if computed_sai != sai:
       reject("ERR_SAI_MISMATCH")

10. Optional: Verify Signature
    if signature present:
        if not verify_signature(sae, signature, public_key):
            reject("ERR_INVALID_SIGNATURE")

11. Commit
    - Store (sai, sae, counter) atomically
    - Update Actor state (last_sai = sai, counter = counter)
    - Return success
```

### 9.2 Verification Pseudocode
```python
def verify_and_commit(
    session_id: str,
    counter: int,
    prev_sai: bytes,
    sae: str,
    sai: bytes
) -> bool:
    # Load state
    k_chain = load_k_chain(session_id)
    actor = load_actor(session_id)
    
    # Verify counter
    if counter != actor.last_counter + 1:
        return False
    
    # Verify prevSAI
    if prev_sai != actor.last_sai:
        return False
    
    # Verify canonicalization
    if not is_canonical(sae):
        return False
    
    # Derive gi
    gi = derive_gi(k_chain, counter)
    
    # Recompute SAI
    computed_sai = compute_sai(prev_sai, sae, gi)
    
    # Verify match
    if computed_sai != sai:
        return False
    
    # Commit atomically
    with transaction():
        store_action(sai, sae)
        actor.last_sai = sai
        actor.last_counter = counter
        save_actor(actor)
    
    return True
```

---

## 10. State Machine

### 10.1 SDK State Transitions
```
[DISCONNECTED]
    |
    | connect()
    ↓
[CONNECTED]
    |
    | sync()
    ↓
[SYNCED]
    |
    | propose(action)
    ↓
[PROPOSING]
    |
    | ← ACK
    ↓
[COMMITTED] → back to SYNCED
    |
    | ← NAK
    ↓
[REJECTED] → resync → SYNCED
    |
    | network error
    ↓
[DISCONNECTED]
```

### 10.2 State Descriptions

**DISCONNECTED:**
- No active session
- Must call `connect()` to establish session

**CONNECTED:**
- Session established
- K_chain received
- Must call `sync()` to fetch latest state

**SYNCED:**
- Counter and prevSAI synchronized
- Ready to propose actions

**PROPOSING:**
- Action submitted
- Waiting for server response
- Counter not yet committed

**COMMITTED:**
- Server ACK received
- Counter and prevSAI updated
- Transition back to SYNCED

**REJECTED:**
- Server NAK received
- Local state may be stale
- Must resync before retrying

---

## 11. Concurrency Model

### 11.1 Single Writer Principle

**One Actor = One serialized history.**

At any given time, only one SDK instance should be actively writing for an Actor.

### 11.2 Backend Locking

Backends SHOULD implement per-Actor locking during commit:
```python
def commit_action(actor_id: str, action: Action):
    with actor_lock(actor_id):
        # Verify counter
        # Verify prevSAI
        # Commit atomically
        pass
```

This prevents:
- Race conditions
- Double-submission
- Counter desync

### 11.3 Multi-Device Scenarios

If multiple devices share the same user:

**Option 1: Separate Actors**
```
device_a: Actor = "user123:device_a"
device_b: Actor = "user123:device_b"
```
Each device has independent history.

**Option 2: Coordinated Write**
```
// Only one device writes at a time
// Other devices read-only or request lock
```

L0 does not define coordination mechanism (out of scope).

---

## 12. Error Codes

### 12.1 Client Errors (4xx)

| Code | Description |
|------|-------------|
| `ERR_INVALID_COUNTER` | Counter is not `last + 1` |
| `ERR_INVALID_PREV_SAI` | prevSAI does not match last SAI |
| `ERR_INVALID_CANONICALIZATION` | SAE violates VAX-JCS rules |
| `ERR_SAI_MISMATCH` | Recomputed SAI does not match submitted SAI |
| `ERR_INVALID_SESSION` | Session not found or expired |
| `ERR_INVALID_SIGNATURE` | Optional signature verification failed |
| `ERR_DUPLICATE_SAI` | SAI already exists (replay attempt) |

### 12.2 Server Errors (5xx)

| Code | Description |
|------|-------------|
| `ERR_STORAGE_FAILURE` | Database or storage error |
| `ERR_INTERNAL` | Unspecified internal error |

### 12.3 Error Response Format
```json
{
  "status": "rejected",
  "error": "ERR_INVALID_COUNTER",
  "detail": "Expected counter 42, got 41",
  "expected": 42,
  "received": 41
}
```

---

## 13. Test Vectors

### 13.1 Canonicalization Test

**Input:**
```json
{
  "name": "Alice",
  "age": 30,
  "city": "Tokyo"
}
```

**Expected Output:**
```json
{"age":30,"city":"Tokyo","name":"Alice"}
```

### 13.2 Genesis SAI Test Vector

**Input:**
- actor_id: `"user123:device456"`
- genesis_salt (hex): `a1a2a3a4a5a6a7a8a9aaabacadaeafb0`

**Expected Genesis SAI (hex):**
```
afc50728cd79e805a8ae06875a1ddf78ca11b0d56ec300b160fb71f50ce658c3
```

**Verification Command:**
```bash
printf '\x56\x41\x58\x2d\x47\x45\x4e\x45\x53\x49\x53'\
       'user123:device456'\
       '\xa1\xa2\xa3\xa4\xa5\xa6\xa7\xa8\xa9\xaa\xab\xac\xad\xae\xaf\xb0' | \
openssl dgst -sha256
```

### 13.3 gi Derivation Test

**Inputs:**
```
K_chain (hex): 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
Counter: 1
```

**Expected Output (hex):**
```
[To be computed by implementation]
```

### 13.4 SAI Computation Test

**Inputs:**
```
prevSAI (hex): 0000000000000000000000000000000000000000000000000000000000000000
SAE: {"action":"test"}
gi (hex): ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
```

**Expected SAI (hex):**
```
[To be computed by implementation]
```

### 13.5 Full Chain Test

**Genesis:**
```
actor_id: "user1:dev1"
genesis_salt (hex): 00112233445566778899aabbccddeeff
SAI_0 (hex): [compute]
```

**Action 1:**
```
counter: 1
prevSAI: SAI_0
SAE: {"action":"deposit","amount":100}
K_chain (hex): [random 32 bytes]
gi_1: [derive]
SAI_1: [compute]
```

**Action 2:**
```
counter: 2
prevSAI: SAI_1
SAE: {"action":"withdraw","amount":50}
gi_2: [derive]
SAI_2: [compute]
```

---

## 14. RFC 2119 Usage

This specification uses RFC 2119 keywords (MUST, SHOULD, etc.) to describe
**reference implementation behavior**, not to impose requirements on other
implementations.

**Interpretation:**

- **MUST** = This is how vax-c and vax-go behave
- **SHOULD** = Recommended behavior for compatibility
- **MAY** = Optional, implementation-specific

Other implementations are free to deviate, but cross-compatibility may be affected.

---

## Appendix A: Wire Format Summary

### A.1 Binary Format (Conceptual)
```
Action Submission:
├─ Headers (HTTP)
│  ├─ X-VAX-Session: UUID
│  ├─ X-VAX-Counter: uint16 (decimal string)
│  └─ X-VAX-PrevSAI: hex(32 bytes)
└─ Body (JSON)
   ├─ sae: canonical JSON string
   ├─ sai: hex(32 bytes)
   └─ signature: base64 (optional)
```

### A.2 Hash Input Structure
```
SAI Computation:
├─ Prefix: "VAX-SAI" (7 bytes ASCII)
├─ prevSAI: 32 bytes (raw binary)
├─ SAE: N bytes (UTF-8 JSON)
└─ gi: 32 bytes (raw binary)

Total: 71 + N bytes
```

### A.3 Storage Schema (Conceptual)
```sql
CREATE TABLE actions (
    sai BINARY(32) PRIMARY KEY,
    actor_id VARCHAR(255) NOT NULL,
    counter INT NOT NULL,
    prev_sai BINARY(32) NOT NULL,
    sae TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_actor_counter (actor_id, counter)
);

CREATE TABLE actor_state (
    actor_id VARCHAR(255) PRIMARY KEY,
    last_sai BINARY(32) NOT NULL,
    last_counter INT NOT NULL,
    genesis_salt BINARY(16) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

---

## Revision History

| Version | Date | Changes |
|---------|------|---------|
| 0.6 | 2025-12-08 | Split from combined doc; added test vectors |
| 0.4 | 2025-11-24 | Added distributed semantics |
| 0.3 | 2025-11-08 | Initial version |

---

**End of Technical Specification**