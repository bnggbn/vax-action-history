# VAX TypeScript SDK

Zero-dependency tamper-evident action history for TypeScript/JavaScript.

## Features

- ✅ **Complete VAX implementation** - JCS, SAE, SDTO, Core primitives
- ✅ **Pure TypeScript** - No external dependencies for core functionality
- ✅ **Cross-platform** - Works in Node.js, Bun, Deno, and browsers
- ✅ **Identical output** - Produces same results as Go and C implementations
- ✅ **Type-safe** - Full TypeScript type definitions

## Installation

```bash
cd ts
npm install
```

## Build

```bash
npm run build
```

## Test

```bash
npm test                # Run all tests
npm run test:watch      # Watch mode
npm run test:coverage   # With coverage
```

---

## Quick Start

### 1. Import the SDK

```typescript
import {
  // Core VAX functions
  computeSAI,
  computeGenesisSAI,
  verifyAction,
  generateGenesisSalt,

  // JCS - JSON Canonicalization
  marshal,

  // SAE - Semantic Action Envelope
  buildSAE,
  signEnvelope,

  // SDTO - Schema-Driven Validation
  newSchemaBuilder,
  newAction,
} from './src';
```

### 2. Define Schema (Backend)

```typescript
import { newSchemaBuilder } from './src';

// Define action constraints
const builder = newSchemaBuilder()
  .setActionStringLength('username', '3', '20')
  .setActionNumberRange('amount', '0', '1000000')
  .setActionEnum('currency', ['USD', 'EUR', 'TWD']);

const schema = builder.buildSchema();

// Transport schema to client via API
const schemaJSON = schema; // Record<string, FieldSpec>
```

### 3. Build Action (Client)

```typescript
import { newAction } from './src';

// Build validated action
const saeBytes = newAction('transfer', schema)
  .set('username', 'alice')
  .set('amount', 500.0)
  .set('currency', 'USD')
  .finalize();

// saeBytes is canonical JSON (JCS)
```

### 4. Compute SAI (Client)

```typescript
import { computeGenesisSAI, computeSAI, generateGenesisSalt } from './src';

// First action: Genesis
const actorID = 'user123:device456';
const genesisSalt = generateGenesisSalt();
const genesisSAI = await computeGenesisSAI(actorID, genesisSalt);

// Subsequent actions
const prevSAI = genesisSAI; // or last SAI from storage
const sai = await computeSAI(prevSAI, saeBytes);

// Submit to backend: { prevSAI, saeBytes, sai }
```

### 5. Verify Action (Backend)

```typescript
import { verifyAction } from './src';

// Backend receives: expectedPrevSAI, prevSAI, saeBytes, clientSAI
const signedEnvelope = await verifyAction(
  expectedPrevSAI,     // What backend expects
  prevSAI,             // What client claims
  saeBytes,            // Client's SAE
  clientSAI,           // Client's computed SAI
  schema,              // Backend's schema
  privateKey           // Backend's Ed25519 private key (64 bytes)
);

// signedEnvelope contains backend signature
// Store: { sai, signedEnvelope, timestamp }
```

---

## API Reference

### Core Functions

#### `computeSAI(prevSAI: Uint8Array, saeBytes: Uint8Array): Promise<Uint8Array>`

Compute SAI for an action.

**Formula:** `SAI_n = SHA256("VAX-SAI" || prevSAI || SHA256(SAE))`

**Example:**
```typescript
const sai = await computeSAI(prevSAI, saeBytes);
// Returns: 32-byte SAI
```

#### `computeGenesisSAI(actorID: string, genesisSalt: Uint8Array): Promise<Uint8Array>`

Compute genesis SAI for an Actor.

**Formula:** `SAI_0 = SHA256("VAX-GENESIS" || actorID || genesisSalt)`

**Example:**
```typescript
const genesisSAI = await computeGenesisSAI('user:device', salt);
// Returns: 32-byte genesis SAI
```

#### `verifyAction(...): Promise<Envelope>`

Backend verification: verify and sign action.

**Returns:** Signed envelope with backend signature

#### `generateGenesisSalt(): Uint8Array`

Generate random 16-byte genesis salt using `crypto.getRandomValues()`.

---

### JCS - JSON Canonicalization Scheme

#### `marshal(value: unknown): Buffer`

Marshal any JavaScript value into canonical JSON bytes.

**Features:**
- UTF-8 encoding
- No whitespace (compact form)
- Lexicographic key ordering
- No scientific notation
- Non-ASCII characters escaped as `\uXXXX`
- NFC normalization

**Example:**
```typescript
const bytes = marshal({ action: 'test', value: 42 });
// Returns: Buffer containing '{"action":"test","value":42}'
```

#### `canonicalizeJSON(input: Buffer | string): Buffer`

Canonicalize existing JSON bytes or string.

#### `normalizeJSONNumber(raw: string): string`

Normalize a JSON number string according to VAX-JCS rules.

---

### SAE - Semantic Action Envelope

#### `buildSAE(actionType: string, sdto: Record<string, unknown>): Buffer`

Build a Semantic Action Envelope using JCS canonicalizer.

**Structure:**
```typescript
{
  action_type: string;
  timestamp: number;
  sdto: Record<string, unknown>;
  signature?: Uint8Array;
}
```

**Example:**
```typescript
const saeBytes = buildSAE('transfer', { username: 'alice', amount: 500 });
```

#### `signEnvelope(envelope: Envelope, privateKey: Uint8Array): Promise<Envelope>`

Sign an envelope with Ed25519 private key.

**Note:** Requires Ed25519 support (Node.js 18+ or `@noble/ed25519` library)

---

### SDTO - Schema-Driven Validation

#### `newSchemaBuilder(): SchemaBuilder`

Create a new schema builder.

**Methods:**
```typescript
builder.setActionStringLength(field: string, min: string, max: string)
builder.setActionNumberRange(field: string, min: string, max: string)
builder.setActionEnum(field: string, values: string[])
builder.setActionSign(field: string, signType: string)
builder.setActionSignMulti(field: string, signTypes: string[])
builder.buildSchema(): Record<string, FieldSpec>
```

#### `newAction(actionType: string, schema: Record<string, FieldSpec>): FluentAction`

Create a new action builder with validation.

**Methods:**
```typescript
action.set(key: string, value: unknown): FluentAction
action.finalize(): Buffer  // Returns JCS-canonicalized SAE bytes
```

**Example:**
```typescript
const saeBytes = newAction('transfer', schema)
  .set('username', 'alice')
  .set('amount', 500)
  .finalize();
```

#### `validateData(data: Record<string, unknown>, schema: Record<string, FieldSpec>): void`

Server-side validation of SDTO against schema.

---

## Core Formulas

```typescript
// Genesis (first action for an Actor)
SAI_0 = SHA256("VAX-GENESIS" || actorID || genesisSalt)

// Subsequent actions
SAI_n = SHA256("VAX-SAI" || prevSAI || SHA256(SAE))

// Where:
// - actorID: "user_id:device_id"
// - genesisSalt: 16 random bytes, persistent per Actor
// - prevSAI: previous action's SAI (or SAI_0)
// - SAE: Semantic Action Encoding (canonical JSON)
```

---

## Key Principles

### Always use JCS

```typescript
// ✅ Correct
import { marshal } from './jcs';
const bytes = marshal(obj);

// ❌ Wrong
const bytes = JSON.stringify(obj);
```

### Validation Happens Early

```typescript
// Validation at .set() time
const action = newAction('transfer', schema)
  .set('amount', -100);  // ❌ Fails immediately if schema disallows negatives

// Errors collected and thrown at .finalize()
const saeBytes = action.finalize();  // Throws if any validations failed
```

### Backend Authority

- Backend **verifies** SAI, never computes it
- Backend **signs** SAE to mark "action entered history"
- Schema is backend's authority
- Backend never repairs or modifies client data (IRP principle)

---

## Test Vectors

The test suite uses **identical test vectors** across all implementations (Go, C, TypeScript) to ensure cross-language compatibility.

Key test categories:
- Basic types (null, boolean, number, string)
- Arrays and objects
- Nested structures
- Unicode handling (including emoji with surrogate pairs)
- Number normalization (including -0 handling)
- Error cases (scientific notation, leading zeros)

---

## Cross-Language Verification

To verify output matches other implementations:

```bash
# Go
cd ../go
go test ./pkg/vax/...

# TypeScript
cd ../ts
npm test
```

All implementations should produce identical canonical JSON output for the same inputs.

---

## Error Handling

```typescript
import {
  VaxError,
  InvalidInputError,
  InvalidPrevSAIError,
  SAIMismatchError
} from './src';

try {
  const sai = await computeSAI(prevSAI, saeBytes);
} catch (error) {
  if (error instanceof InvalidInputError) {
    // Handle invalid input
  } else if (error instanceof SAIMismatchError) {
    // Handle SAI verification failure
  }
}
```

---

## Platform Compatibility

### Node.js
```bash
node --version  # 18+ recommended for Ed25519 support
npm test
```

### Bun
```bash
bun test
```

### Deno
```typescript
import { computeSAI } from './src/index.ts';
```

### Browser
The SDK works in modern browsers that support:
- `crypto.subtle` (Web Crypto API)
- `TextEncoder` / `TextDecoder`
- ES2020+

---

## License

MIT License — Free to use, modify, and distribute with attribution.

---

## Documentation

- [Architecture & Design Philosophy](../docs/ARCHITECTURE.md)
- [L0 Technical Specification](../docs/SPECIFICATION.md)
- [Go API Reference](../go/README.md)

---

## Contributing

Cross-language test vectors, semantic edge cases, and tooling improvements are welcome.

VAX grows by **usage**, not mandates.
