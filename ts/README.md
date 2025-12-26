# VAX JCS - TypeScript Implementation

TypeScript implementation of VAX JSON Canonicalization Scheme (JCS).

## Features

- ✅ UTF-8 encoding
- ✅ No whitespace (compact form)
- ✅ Lexicographic key ordering
- ✅ No scientific notation
- ✅ Non-ASCII characters escaped as `\uXXXX`
- ✅ NFC normalization support
- ✅ Identical output to Go implementation

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

## Usage

```typescript
import { marshal, canonicalizeJSON } from 'vax-jcs';

// Marshal any value
const obj = { name: 'Alice', age: 30 };
const canonical = marshal(obj);
console.log(canonical.toString());
// Output: {"age":30,"name":"Alice"}

// Canonicalize existing JSON
const json = '{"z": 1, "a": 2}';
const result = canonicalizeJSON(json);
console.log(result.toString());
// Output: {"a":2,"z":1}
```

## API

### `marshal(value: unknown): Buffer`

Marshal any JavaScript value into canonical JSON bytes.

**Example:**
```typescript
const bytes = marshal({ action: 'test', value: 42 });
// Returns: Buffer containing '{"action":"test","value":42}'
```

### `canonicalizeJSON(input: Buffer | string): Buffer`

Canonicalize existing JSON bytes or string.

**Example:**
```typescript
const canonical = canonicalizeJSON('{ "b": 2, "a": 1 }');
// Returns: Buffer containing '{"a":1,"b":2}'
```

### `canonicalizeValue(value: unknown): Buffer`

Alias for `marshal()`.

### `normalizeJSONNumber(raw: string): string`

Normalize a JSON number string according to VAX-JCS rules.

## Test Vectors

The test suite uses **identical test vectors** to the Go implementation to ensure cross-language compatibility.

Key test categories:
- Basic types (null, boolean, number, string)
- Arrays and objects
- Nested structures
- Unicode handling (including emoji with surrogate pairs)
- Number normalization (including -0 handling)
- Error cases (scientific notation, leading zeros)

## Compliance

This implementation follows the VAX-JCS specification:
- No scientific notation allowed
- No leading zeros (except `0` itself)
- `-0` normalized to `0`
- Non-ASCII characters escaped as UTF-16 code units
- Keys sorted lexicographically
- Compact output (no whitespace)

## Cross-Language Verification

To verify output matches Go implementation:

```bash
# Go
cd ../go
go test ./internal/jcs -v

# TypeScript
cd ../ts
npm test
```

Both should produce identical canonical JSON output for the same inputs.
