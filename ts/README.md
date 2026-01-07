# VAX TypeScript SDK

Pure TypeScript implementation of VAX cryptographic primitives.

Zero dependencies. Works in Node.js, Bun, Deno, and browsers.

---

## Installation

```bash
npm install vax
```

---

## Philosophy

VAX provides **primitives**, not a complete system.

**You control:**
- Storage structure
- Whether to add signatures
- Authorization logic
- Transport protocol

**VAX provides:**
- SAI chain computation
- JCS canonicalization
- Schema validation

---

## Quick Start

```typescript
import { computeGenesisSAI, computeSAI, generateGenesisSalt } from 'vax';

// 1. Genesis
const actorID = 'user123:device456';
const genesisSalt = generateGenesisSalt();
const genesisSAI = await computeGenesisSAI(actorID, genesisSalt);

// 2. Build action
const saeBytes = buildAction('transfer', data);

// 3. Compute SAI
const sai = await computeSAI(genesisSAI, saeBytes);

// 4. Store (your structure)
store({ sai, sae: saeBytes, prevSAI: genesisSAI });
```

---

## Core API

### Genesis

```typescript
async function computeGenesisSAI(
  actorID: string,
  genesisSalt: Uint8Array
): Promise<Uint8Array>
```

Compute genesis SAI for an Actor.

**Formula:**
```
SAI_0 = SHA256("VAX-GENESIS" || actorID || genesisSalt)
```

**Example:**
```typescript
const actorID = 'user123:device456';
const salt = generateGenesisSalt();  // 16 random bytes
const sai = await computeGenesisSAI(actorID, salt);
```

---

### Chain

```typescript
async function computeSAI(
  prevSAI: Uint8Array,
  saeBytes: Uint8Array
): Promise<Uint8Array>
```

Compute SAI for an action.

**Formula:**
```
SAI_n = SHA256("VAX-SAI" || prevSAI || SHA256(SAE))
```

**Example:**
```typescript
const sae = buildSAE('transfer', data);
const sai = await computeSAI(prevSAI, sae);
```

---

### Verification

```typescript
async function verifyChain(
  expectedPrevSAI: Uint8Array,
  saeBytes: Uint8Array,
  clientSAI: Uint8Array
): Promise<void>
```

Verify SAI chain integrity.

**Throws:**
- `InvalidPrevSAIError` - Chain discontinuity
- `SAIMismatchError` - SAI computation mismatch

**Example:**
```typescript
try {
  await verifyChain(expectedPrevSAI, saeBytes, clientSAI);
  // Chain valid
} catch (error) {
  if (error instanceof SAIMismatchError) {
    // SAI mismatch
  }
}
```

---

## Schema-Driven Validation (SDTO)

### Define Schema

```typescript
import { newSchemaBuilder } from 'vax';

const schema = newSchemaBuilder()
  .setActionStringLength('username', '3', '20')
  .setActionNumberRange('amount', '0', '1000000')
  .setActionEnum('currency', ['USD', 'EUR', 'TWD'])
  .buildSchema();
```

### Build Validated Action

```typescript
import { newAction } from 'vax';

const saeBytes = newAction('transfer', schema)
  .set('username', 'alice')
  .set('amount', 500.0)
  .set('currency', 'USD')
  .finalize();  // Returns Buffer with canonical JSON
```

**Validation happens at `.set()` time.** Errors collected and thrown at `.finalize()`.

### Server-Side Validation

```typescript
import { validateData } from 'vax';

// Backend validates SDTO against schema
try {
  validateData(action.sdto, schema);
} catch (error) {
  // Schema violation
}
```

---

## JCS (JSON Canonicalization)

```typescript
import { marshal } from 'vax';

// Marshal to canonical JSON
const canonical = marshal(obj);

// Always produces identical output
const obj1 = { b: 2, a: 1 };
const obj2 = { a: 1, b: 2 };

const bytes1 = marshal(obj1);  // {"a":1,"b":2}
const bytes2 = marshal(obj2);  // {"a":1,"b":2}
// bytes1.equals(bytes2) ✅
```

**Never use `JSON.stringify()` for SAE. Always use `marshal()`.**

---

## SAE (Semantic Action Envelope)

```typescript
import { buildSAE } from 'vax';

const saeBytes = buildSAE('transfer', {
  username: 'alice',
  amount: 500.0,
  currency: 'USD',
});

// Returns Buffer with canonical JSON:
// {"action_type":"transfer","sdto":{...},"timestamp":1704672000000}
```

**Structure:**
```typescript
interface Envelope {
  action_type: string;
  timestamp: number;
  sdto: Record<string, unknown>;
}
```

**No signature field.** If you need signatures, add them yourself.

---

## Complete Workflow

### Client Side

```typescript
import { computeSAI, newAction } from 'vax';

// 1. Get schema from backend
const schema = await backend.getSchema('transfer');

// 2. Build validated action
const saeBytes = newAction('transfer', schema)
  .set('username', 'alice')
  .set('amount', 500.0)
  .set('currency', 'USD')
  .finalize();

// 3. Compute SAI
const prevSAI = getLastSAI();
const sai = await computeSAI(prevSAI, saeBytes);

// 4. Submit to backend
await backend.submit({
  sai,
  sae: saeBytes,
  prevSAI,
});
```

### Backend Side

```typescript
import { verifyChain, validateData } from 'vax';

// 1. Verify chain
await verifyChain(expectedPrevSAI, req.sae, req.sai);

// 2. Validate schema
const env = JSON.parse(req.sae.toString());
validateData(env.sdto, schema);

// 3. Optional: Add signature (using @noble/ed25519 or similar)
const signature = await ed25519.sign(req.sae, privateKey);

// 4. Store (your structure)
await db.store({
  sai: req.sai,
  sae: req.sae,
  prevSAI: req.prevSAI,
  signature,  // optional
  timestamp: Date.now(),
});
```

---

## Optional: Signatures

VAX doesn't handle signatures. Use standard libraries:

### Node.js 18+ (Web Crypto)

```typescript
// Note: Ed25519 support requires Node.js 18+ with experimental flag
// For production, use @noble/ed25519

const keyPair = await crypto.subtle.generateKey(
  'Ed25519',
  true,
  ['sign', 'verify']
);

const signature = await crypto.subtle.sign(
  'Ed25519',
  keyPair.privateKey,
  saeBytes
);
```

### Using @noble/ed25519

```typescript
import * as ed25519 from '@noble/ed25519';

// Generate key pair
const privateKey = ed25519.utils.randomPrivateKey();
const publicKey = await ed25519.getPublicKey(privateKey);

// Sign
const signature = await ed25519.sign(saeBytes, privateKey);

// Verify
const valid = await ed25519.verify(signature, saeBytes, publicKey);
```

**Storage is your choice:**

```typescript
// Option A: Separate field
interface ActionRecord {
  sai: Uint8Array;
  sae: Uint8Array;
  prevSAI: Uint8Array;
  signature?: Uint8Array;  // Independent
}

// Option B: Wrap in metadata
interface ActionWithMeta {
  action: {
    sai: Uint8Array;
    sae: Uint8Array;
    prevSAI: Uint8Array;
  };
  metadata: {
    signature?: Uint8Array;
    timestamp: number;
    author: string;
  };
}

// Your choice!
```

---

## Helpers

```typescript
// Generate random genesis salt
const salt = generateGenesisSalt();  // 16 bytes

// Hex encoding
const hex = toHex(sai);
const sai = fromHex(hex);

// Constants
SAI_SIZE           // 32
GENESIS_SALT_SIZE  // 16
```

---

## Error Handling

```typescript
import {
  VaxError,
  InvalidInputError,
  InvalidPrevSAIError,
  SAIMismatchError
} from 'vax';

try {
  await computeSAI(prevSAI, saeBytes);
} catch (error) {
  if (error instanceof InvalidInputError) {
    // Invalid input parameters
  } else if (error instanceof SAIMismatchError) {
    // SAI computation mismatch
  }
}
```

---

## Platform Compatibility

### Node.js

```bash
node --version  # 18+ recommended
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

Works in modern browsers supporting:
- `crypto.subtle` (Web Crypto API)
- `TextEncoder` / `TextDecoder`
- ES2020+

---

## Testing

```bash
npm test              # Run all tests
npm run test:watch   # Watch mode
npm run test:coverage # With coverage
```

---

## Examples

See [examples/](examples/) for complete workflows:

- **[minimal](examples/minimal/)** - Simplest usage, no signatures
- **[with-signature](examples/with-signature/)** - Add backend signature
- **[browser](examples/browser/)** - Browser usage
- **[node](examples/node/)** - Node.js usage

---

## Key Principles

### 1. Always use marshal()

```typescript
// ✅ Correct
import { marshal } from 'vax';
const bytes = marshal(obj);

// ❌ Wrong
const bytes = Buffer.from(JSON.stringify(obj));
```

### 2. Validate early

```typescript
// Validation at .set() time
const action = newAction('transfer', schema)
  .set('amount', -100);  // ❌ Fails immediately

// Errors collected at .finalize()
const saeBytes = action.finalize();
```

### 3. Backend verifies, never repairs

```typescript
// ✅ Backend verifies
await verifyChain(...);
validateData(...);

// ❌ Backend never "fixes" client data
// Don't do: data.amount = Math.abs(data.amount)
```

---

## Performance

Pure TypeScript with no dependencies:

```
ComputeSAI:        ~25 µs
ComputeGenesisSAI: ~12 µs
JCS Marshal:       ~8 µs
```

---

## Cross-Language Compatibility

All implementations produce identical output:

```bash
# Test vector
actorID: "user123:device456"
genesisSalt: a1a2a3a4a5a6a7a8a9aaabacadaeafb0

# Expected
genesisSAI: afc50728cd79e805a8ae06875a1ddf78ca11b0d56ec300b160fb71f50ce658c3

# Verify
npm test
```

---

## Documentation

- [Architecture](../docs/ARCHITECTURE.md)
- [L0 Specification](../docs/SPECIFICATION.md)

---

## License

MIT License

---

## Philosophy

> VAX is a tool, not a system.
>
> We provide primitives.
> You decide how to use them.
