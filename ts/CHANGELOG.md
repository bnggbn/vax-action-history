# CHANGELOG

## [Unreleased]

### Added
- Pure TypeScript implementation of VAX L0
- JCS (JSON Canonicalization Scheme) implementation with cross-language test vectors

### Changed
- N/A

### Fixed
- N/A

---

## 2026-01-06

### Added
- **SDTO SDK** (`src/sdto/`)
  - `FieldSpec` interface: validation rule definition (type, min, max, enum)
  - `FluentAction` class: fluent builder with chained `.set()` calls and instant validation
  - `SchemaBuilder` class: schema definition builder for providers
  - `parseSchema()`: cross-service deserialization from `Record<string, unknown>` to `Record<string, FieldSpec>`
  - Factory functions: `newAction()`, `newSchemaBuilder()` to match Go naming conventions

### Changed
- **Validation logic**
  - String length validation uses numeric parsing (`parseInt()`)
  - Number validation with min/max boundary checks
  - Enum validation with strict equality

### Notes
- This implementation maintains identical validation behavior with the Go SDK
- All validation errors are accumulated and thrown on `finalize()`

---

## 2026-01-07

### Removed
- **Sign type support in SDTO** (reverted)
  - Removed `"sign"` type from `FieldSpec` interface
  - Removed `SupportedSignTypes` constant from `SchemaBuilder`
  - Removed `setActionSign()` and `setActionSignMulti()` methods
  - Removed `validateSign()` function from `FluentAction`
  - Updated exports in `sdto/index.ts`

- **SAE signing functionality**
  - Removed `signature?: Uint8Array` field from `Envelope` interface
  - Removed `signEnvelope()` function
  - Removed `generateKeyPair()` function

- **VerifyAction signing**
  - Removed `privateKey` parameter from `verifyAction()`
  - Removed signature verification checks
  - Removed `signEnvelope` import from `vax.ts`

### Notes
- Signature functionality removed to simplify the SDK architecture
- Focus on core tamper-evident action history without embedded signing
  - `SupportedSignTypes` constant: `["ed25519", "rsa", "ecdsa"]`
  - `SchemaBuilder.setActionSign()`: Set single signature algorithm
  - `SchemaBuilder.setActionSignMulti()`: Set multiple allowed signature algorithms

- **Enhanced VAX Core**
  - `verifyAction()`: Full verification (crypto + schema validation, matches Go signature)
  - `verifyPrevSAI()`: Simple prevSAI verification (renamed from old `verifyAction`)
  - `SAIMismatchError`: New error class for SAI verification failures

### Changed
- **BREAKING: `computeSAI()` now deterministic**
  - Removed internal random `gi` generation
  - Formula: `SHA256("VAX-SAI" || prevSAI || SHA256(SAE))`
  - Same inputs now produce same outputs (matches Go version)

- **BREAKING: `FluentAction.finalize()` signature change**
  - Old: Returns `{ actionType: string; data: Record<string, unknown> }`
  - New: Returns `Buffer` (JCS-canonicalized SAE bytes)
  - Automatically calls `buildSAE()` internally

- **BREAKING: Removed exports**
  - `GI_SIZE` constant (gi no longer used in computeSAI)
  - Old `verifyAction()` renamed to `verifyPrevSAI()`

### Migration Guide

**computeSAI() changes:**
```typescript
// Before (non-deterministic with random gi)
const sai1 = await computeSAI(prevSAI, saeBytes);
const sai2 = await computeSAI(prevSAI, saeBytes);
// sai1 !== sai2 (different due to random gi)

// After (deterministic)
const sai1 = await computeSAI(prevSAI, saeBytes);
const sai2 = await computeSAI(prevSAI, saeBytes);
// sai1 === sai2 (same inputs → same output)
```

**FluentAction.finalize() changes:**
```typescript
// Before
const result = action.set("name", "Alice").finalize();
console.log(result.actionType, result.data);

// After
const saeBytes = action.set("name", "Alice").finalize();
// saeBytes is JCS-canonicalized Buffer ready for hashing
const sai = await computeSAI(prevSAI, saeBytes);
```

**verifyAction() renamed:**
```typescript
// Before
verifyAction(expectedPrevSAI, prevSAI);

// After
verifyPrevSAI(expectedPrevSAI, prevSAI);

// Or use full verification with schema
await verifyAction(expectedPrevSAI, prevSAI, saeBytes, clientSAI, schema, privateKey);
```

### Notes
- TypeScript implementation now fully matches Go SDK architecture
- All changes maintain cross-language compatibility
- Ed25519 signing requires Node.js 18+ or modern browsers with Web Crypto support
- For broader compatibility, consider using `@noble/ed25519` library
