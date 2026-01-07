# Migration Guide: v0.7 → v0.8

VAX v0.8 simplifies the API by removing built-in signature handling.

**Philosophy change:** Tool, not System.

---

## TL;DR

### What Changed

```diff
- VAX manages signatures
+ You manage signatures (if needed)

- Envelope has signature field
+ Envelope is signature-free

- VerifyAction returns signed envelope
+ VerifyChain returns error only

- Backend must provide private key
+ Backend decides whether to sign
```

---

## Breaking Changes

### 1. Envelope Structure

#### Before (v0.7)

```go
type Envelope struct {
    ActionType string         `json:"action_type"`
    Timestamp  int64          `json:"timestamp"`
    SDTO       map[string]any `json:"sdto"`
    Signature  []byte         `json:"signature"`  // ❌ Removed
}
```

#### After (v0.8)

```go
type Envelope struct {
    ActionType string         `json:"action_type"`
    Timestamp  int64          `json:"timestamp"`
    SDTO       map[string]any `json:"sdto"`
    // No signature field
}
```

**Migration:**

If you stored Envelope with signature, extract signature to separate field:

```go
// Migration function
func MigrateV7ToV8(oldData []byte) (ActionRecord, error) {
    var oldEnv struct {
        ActionType string         `json:"action_type"`
        Timestamp  int64          `json:"timestamp"`
        SDTO       map[string]any `json:"sdto"`
        Signature  []byte         `json:"signature"`
    }
    json.Unmarshal(oldData, &oldEnv)

    // Create new envelope without signature
    newEnv := Envelope{
        ActionType: oldEnv.ActionType,
        Timestamp:  oldEnv.Timestamp,
        SDTO:       oldEnv.SDTO,
    }

    // Store signature separately
    return ActionRecord{
        SAI:       computedSAI,  // SAI doesn't change!
        SAE:       jcs.Marshal(newEnv),
        Signature: oldEnv.Signature,
    }, nil
}
```

---

### 2. VerifyAction API

#### Before (v0.7)

```go
func VerifyAction(
    expectedPrevSAI []byte,
    prevSAI []byte,
    saeBytes []byte,
    clientSAI []byte,
    schema map[string]FieldSpec,
    privateKey ed25519.PrivateKey,  // ❌ Removed
) (*Envelope, error)  // ❌ Returns signed envelope
```

#### After (v0.8)

```go
func VerifyChain(
    expectedPrevSAI []byte,
    saeBytes []byte,
    clientSAI []byte,
) error  // ✅ Returns error only
```

**Migration:**

```go
// Before (v0.7)
signedEnv, err := vax.VerifyAction(
    expectedPrevSAI,
    prevSAI,
    saeBytes,
    clientSAI,
    schema,
    privateKey,  // Backend private key
)
if err != nil {
    return err
}
db.Store(signedEnv)

// After (v0.8)
err := vax.VerifyChain(
    expectedPrevSAI,
    saeBytes,
    clientSAI,
)
if err != nil {
    return err
}

// Validate schema separately
var env Envelope
json.Unmarshal(saeBytes, &env)
err = sdto.ValidateData(env.SDTO, schema)
if err != nil {
    return err
}

// Sign yourself if needed
signature := ed25519.Sign(privateKey, saeBytes)

// Store with your structure
db.Store(ActionRecord{
    SAI:       clientSAI,
    SAE:       saeBytes,
    PrevSAI:   expectedPrevSAI,
    Signature: signature,  // Optional
})
```

---

### 3. Removed Functions

#### v0.7 Functions (Removed)

```go
❌ func (env *Envelope) Sign(privateKey ed25519.PrivateKey) error
❌ func SignEnvelope(env *Envelope, privateKey []byte) error
❌ func GenerateKeyPair() ([]byte, []byte, error)
❌ func VerifySignature(envelope *Envelope, publicKey []byte) bool
```

#### Use Standard Libraries Instead

```go
import "crypto/ed25519"

// Generate key pair
publicKey, privateKey, err := ed25519.GenerateKey(nil)

// Sign
signature := ed25519.Sign(privateKey, saeBytes)

// Verify
valid := ed25519.Verify(publicKey, saeBytes, signature)
```

---

### 4. Storage Structure Changes

#### Before (v0.7)

```go
// Signature embedded in SAE
{
    "sai": "...",
    "sae": {
        "action_type": "transfer",
        "timestamp": 123,
        "sdto": {...},
        "signature": "..."  // ❌ Embedded
    }
}
```

**Problem:** SAI computed from unsigned SAE, but stored SAE has signature → Merkle verification fails!

#### After (v0.8)

```go
// Signature separate
{
    "sai": "...",
    "sae": {
        "action_type": "transfer",
        "timestamp": 123,
        "sdto": {...}
        // No signature field
    },
    "signature": "..."  // ✅ Separate (optional)
}
```

**Benefit:** SAI verification works correctly!

---

## Migration Steps

### Step 1: Update Dependencies

```bash
# Go
go get github.com/bnggbn/vax-action-history/go/pkg/vax@v0.8.0

# TypeScript
npm install vax@0.8.0

# C
# Update submodule or download v0.8.0 release
```

---

### Step 2: Update Code

#### Client Code

```go
// Before (v0.7)
saeBytes := sae.BuildSAE("transfer", data)
sai := vax.ComputeSAI(prevSAI, saeBytes)
backend.Submit(sai, saeBytes, prevSAI)

// After (v0.8) - No change for client!
saeBytes := sae.BuildSAE("transfer", data)
sai := vax.ComputeSAI(prevSAI, saeBytes)
backend.Submit(sai, saeBytes, prevSAI)
```

**Client code unchanged!** SAE structure is the same (no signature in v0.8).

#### Backend Code

```go
// Before (v0.7)
signedEnv, err := vax.VerifyAction(
    expectedPrevSAI,
    prevSAI,
    saeBytes,
    clientSAI,
    schema,
    privateKey,
)

// After (v0.8)
err := vax.VerifyChain(expectedPrevSAI, saeBytes, clientSAI)
if err != nil {
    return err
}

// Schema validation (separate)
var env sae.Envelope
json.Unmarshal(saeBytes, &env)
err = sdto.ValidateData(env.SDTO, schema)
if err != nil {
    return err
}

// Optional: Add signature
var signature []byte
if needsSignature {
    signature = ed25519.Sign(privateKey, saeBytes)
}

// Store
db.Store(ActionRecord{
    SAI:       clientSAI,
    SAE:       saeBytes,
    PrevSAI:   expectedPrevSAI,
    Signature: signature,
})
```

---

### Step 3: Migrate Data

If you have existing data with signatures embedded in SAE:

```go
func MigrateDatabase() error {
    actions := db.GetAllActions()

    for _, old := range actions {
        // Parse old SAE (with signature)
        var oldEnv struct {
            ActionType string         `json:"action_type"`
            Timestamp  int64          `json:"timestamp"`
            SDTO       map[string]any `json:"sdto"`
            Signature  []byte         `json:"signature,omitempty"`
        }
        json.Unmarshal(old.SAE, &oldEnv)

        // Extract signature
        signature := oldEnv.Signature

        // Create new SAE (without signature)
        newEnv := sae.Envelope{
            ActionType: oldEnv.ActionType,
            Timestamp:  oldEnv.Timestamp,
            SDTO:       oldEnv.SDTO,
        }
        newSAE := jcs.Marshal(newEnv)

        // SAI stays the same!
        // (Because client computed SAI from unsigned SAE)

        // Store in new format
        db.StoreV8(ActionRecord{
            SAI:       old.SAI,
            SAE:       newSAE,
            PrevSAI:   old.PrevSAI,
            Signature: signature,  // Now separate
        })
    }

    return nil
}
```

---

### Step 4: Update Tests

```go
// Before (v0.7)
func TestVerifyAction(t *testing.T) {
    privateKey := generateKey()

    signedEnv, err := vax.VerifyAction(..., privateKey)
    assert.NoError(t, err)
    assert.NotNil(t, signedEnv.Signature)
}

// After (v0.8)
func TestVerifyChain(t *testing.T) {
    err := vax.VerifyChain(expectedPrevSAI, saeBytes, clientSAI)
    assert.NoError(t, err)

    // If you need signatures, test separately
    signature := ed25519.Sign(privateKey, saeBytes)
    valid := ed25519.Verify(publicKey, saeBytes, signature)
    assert.True(t, valid)
}
```

---

## Why This Change?

### Problem in v0.7

```
SAI = SHA256("VAX-SAI" || prevSAI || SHA256(SAE))

But if Signature is in SAE:
  Client computes SAI with unsigned SAE
  Backend adds signature to SAE
  Stored SAE has signature
  → SHA256(stored SAE) ≠ SAI
  → Merkle verification broken!
```

### Solution in v0.8

```
Remove Signature from Envelope structure
→ SAE is always signature-free
→ SAI stable
→ Merkle verification works
→ Signatures stored separately (if needed)
```

### Philosophy Shift

```
v0.7: VAX is a system
  - VAX manages signatures
  - VAX enforces structure
  - One way to do things

v0.8: VAX is a tool
  - VAX provides primitives
  - You decide structure
  - Multiple patterns supported
```

---

## Benefits of v0.8

### 1. No Circular Dependency

```
v0.7: Signature in SAE → SAI includes signature → Chicken-egg problem
v0.8: Signature separate → SAI stable → Clean design
```

### 2. Flexibility

```
v0.7: Must use Ed25519, must provide private key
v0.8: Use any signature scheme (Ed25519, RSA, ECDSA...)
     or don't sign at all
```

### 3. Simplicity

```
v0.7: 1200 lines (including signature handling)
v0.8: ~800 lines (core primitives only)
```

### 4. Clear Responsibility

```
VAX provides: Integrity primitives
You provide: Signatures, storage, authorization
```

---

## FAQ

### Q: Do I need to resign all my data?

**A:** No! SAI stays the same because clients computed SAI from unsigned SAE.

You just need to extract the signature from SAE and store it separately.

### Q: What if I don't need signatures?

**A:** Perfect! Just don't add them. v0.8 makes signatures optional.

### Q: Can I still use Ed25519?

**A:** Yes! Use `crypto/ed25519` from Go standard library, or `@noble/ed25519` for TypeScript.

### Q: Will SAI change after migration?

**A:** No! SAI is computed from unsigned SAE, which doesn't change.

### Q: Is v0.8 backward compatible?

**A:** No, it's a breaking change. But migration is straightforward (see above).

### Q: Should I migrate now?

**A:** If you have production data, test migration carefully. If you're starting new, use v0.8.

---

## Checklist

- [ ] Update dependencies to v0.8
- [ ] Remove `Signature` field from `Envelope` struct
- [ ] Change `VerifyAction()` to `VerifyChain()`
- [ ] Remove `privateKey` parameter from backend verification
- [ ] Add signature handling externally (if needed)
- [ ] Update storage structure (signature separate)
- [ ] Migrate existing data
- [ ] Update tests
- [ ] Update documentation

---

## Support

Questions about migration? Open an issue on GitHub or check the docs:

- [v0.8 Release Notes](https://github.com/bnggbn/vax-action-history/releases/tag/v0.8.0)
- [Architecture](../docs/ARCHITECTURE.md)
- [Examples](../examples/)

---

**Remember:** v0.8 makes VAX simpler and more flexible. You're in control.
