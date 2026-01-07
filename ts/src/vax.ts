/**
 * VAX TypeScript SDK
 *
 * Git-like tamper-evident action history with deterministic output.
 */

import { FieldSpec, validateData } from './sdto';
import { Envelope } from './sae';

// ============================================================================
// Constants
// ============================================================================

export const SAI_SIZE = 32;
export const GENESIS_SALT_SIZE = 16;

// ============================================================================
// Error Classes
// ============================================================================

export class VaxError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'VaxError';
  }
}

export class InvalidInputError extends VaxError {
  constructor(message = 'invalid input') {
    super(message);
    this.name = 'InvalidInputError';
  }
}

export class InvalidPrevSAIError extends VaxError {
  constructor(message = 'invalid prevSAI') {
    super(message);
    this.name = 'InvalidPrevSAIError';
  }
}

export class SAIMismatchError extends VaxError {
  constructor(message = 'SAI mismatch') {
    super(message);
    this.name = 'SAIMismatchError';
  }
}

// ============================================================================
// Internal Helpers
// ============================================================================

/**
 * Constant-time byte array comparison
 */
function bytesEqual(a: Uint8Array, b: Uint8Array): boolean {
  if (a.length !== b.length) {
    return false;
  }
  let result = 0;
  for (let i = 0; i < a.length; i++) {
    result |= a[i] ^ b[i];
  }
  return result === 0;
}

/**
 * Concatenate multiple Uint8Arrays
 */
function concat(...arrays: Uint8Array[]): Uint8Array {
  const totalLength = arrays.reduce((sum, arr) => sum + arr.length, 0);
  const result = new Uint8Array(totalLength);
  let offset = 0;
  for (const arr of arrays) {
    result.set(arr, offset);
    offset += arr.length;
  }
  return result;
}

/**
 * Convert string to Uint8Array (UTF-8)
 */
function stringToBytes(str: string): Uint8Array {
  return new TextEncoder().encode(str);
}

/**
 * SHA-256 hash using Web Crypto API
 */
async function sha256(data: Uint8Array): Promise<Uint8Array> {
  const hashBuffer = await crypto.subtle.digest('SHA-256', data);
  return new Uint8Array(hashBuffer);
}

// ============================================================================
// Public API
// ============================================================================

/**
 * Compute SAI_n = SHA256("VAX-SAI" || prevSAI || SHA256(SAE))
 *
 * Matches Go's ComputeSAI signature (no gi parameter).
 *
 * @param prevSAI - Previous SAI (32 bytes)
 * @param saeBytes - Semantic Action Envelope bytes (must be JCS-canonicalized)
 * @returns Promise<Uint8Array> - New SAI (32 bytes)
 */
export async function computeSAI(
  prevSAI: Uint8Array,
  saeBytes: Uint8Array
): Promise<Uint8Array> {
  if (prevSAI.length !== SAI_SIZE) {
    throw new InvalidInputError(`prevSAI must be ${SAI_SIZE} bytes, got ${prevSAI.length}`);
  }
  if (saeBytes.length === 0) {
    throw new InvalidInputError('saeBytes cannot be empty');
  }

  // Two-stage hash
  const saeHash = await sha256(saeBytes);

  // message = "VAX-SAI" || prevSAI || saeHash
  const label = stringToBytes('VAX-SAI');
  const message = concat(label, prevSAI, saeHash);

  return sha256(message);
}

/**
 * Compute genesis SAI_0 = SHA256("VAX-GENESIS" || actor_id || genesis_salt)
 *
 * @param actorID - Actor identifier string
 * @param genesisSalt - Random salt for genesis (16 bytes)
 * @returns Promise<Uint8Array> - Genesis SAI (32 bytes)
 */
export async function computeGenesisSAI(
  actorID: string,
  genesisSalt: Uint8Array
): Promise<Uint8Array> {
  if (genesisSalt.length !== GENESIS_SALT_SIZE) {
    throw new InvalidInputError(`genesisSalt must be ${GENESIS_SALT_SIZE} bytes, got ${genesisSalt.length}`);
  }

  // message = "VAX-GENESIS" || actorID || genesisSalt
  const label = stringToBytes('VAX-GENESIS');
  const actorIDBytes = stringToBytes(actorID);
  const message = concat(label, actorIDBytes, genesisSalt);

  return sha256(message);
}

/**
 * Verify an action submission (crypto + schema validation)
 * saeBytes: canonical JSON bytes from client (already JCS-marshaled by Finalize)
 *
 * Matches Go's VerifyAction signature.
 *
 * @param expectedPrevSAI - Expected previous SAI (32 bytes)
 * @param prevSAI - Submitted previous SAI (32 bytes)
 * @param saeBytes - SAE bytes (JCS-canonicalized)
 * @param clientProvidedSAI - Client-computed SAI (32 bytes)
 * @param schema - Validation schema
 * @returns Promise<Envelope> - Envelope
 */
export async function verifyAction(
  expectedPrevSAI: Uint8Array,
  prevSAI: Uint8Array,
  saeBytes: Uint8Array,
  clientProvidedSAI: Uint8Array,
  schema: Record<string, FieldSpec>
): Promise<Envelope> {
  // Input validation
  if (expectedPrevSAI.length !== SAI_SIZE) {
    throw new InvalidInputError(`expectedPrevSAI must be ${SAI_SIZE} bytes`);
  }
  if (prevSAI.length !== SAI_SIZE) {
    throw new InvalidInputError(`prevSAI must be ${SAI_SIZE} bytes`);
  }
  if (saeBytes.length === 0) {
    throw new InvalidInputError('saeBytes cannot be empty');
  }

  // Parse SAE from bytes
  let envelope: Envelope;
  try {
    const json = new TextDecoder().decode(saeBytes);
    envelope = JSON.parse(json) as Envelope;
  } catch {
    throw new InvalidInputError('invalid SAE JSON');
  }

  // Verify prevSAI matches
  if (!bytesEqual(prevSAI, expectedPrevSAI)) {
    throw new InvalidPrevSAIError();
  }

  // Verify SDTO against schema
  validateData(envelope.sdto, schema);

  // Verify clientProvidedSAI length
  if (clientProvidedSAI.length !== SAI_SIZE) {
    throw new InvalidInputError(`clientProvidedSAI must be ${SAI_SIZE} bytes`);
  }

  // Verify SAI
  const computedSAI = await computeSAI(prevSAI, saeBytes);
  if (!bytesEqual(computedSAI, clientProvidedSAI)) {
    throw new SAIMismatchError();
  }

  return envelope;
}

/**
 * Simple verification (crypto only, no JSON validation)
 * For cases where you just need to verify prevSAI matches.
 *
 * @param expectedPrevSAI - Expected previous SAI (32 bytes)
 * @param prevSAI - Submitted previous SAI (32 bytes)
 * @throws InvalidInputError if inputs are invalid
 * @throws InvalidPrevSAIError if prevSAI doesn't match
 */
export function verifyPrevSAI(
  expectedPrevSAI: Uint8Array,
  prevSAI: Uint8Array
): void {
  if (expectedPrevSAI.length !== SAI_SIZE) {
    throw new InvalidInputError(`expectedPrevSAI must be ${SAI_SIZE} bytes, got ${expectedPrevSAI.length}`);
  }
  if (prevSAI.length !== SAI_SIZE) {
    throw new InvalidInputError(`prevSAI must be ${SAI_SIZE} bytes, got ${prevSAI.length}`);
  }

  // Verify prevSAI matches
  if (!bytesEqual(prevSAI, expectedPrevSAI)) {
    throw new InvalidPrevSAIError();
  }
}

// ============================================================================
// Utility Functions
// ============================================================================

/**
 * Convert Uint8Array to hex string
 */
export function toHex(bytes: Uint8Array): string {
  return Array.from(bytes)
    .map(b => b.toString(16).padStart(2, '0'))
    .join('');
}

/**
 * Convert hex string to Uint8Array
 */
export function fromHex(hex: string): Uint8Array {
  if (hex.length % 2 !== 0) {
    throw new InvalidInputError('hex string must have even length');
  }
  const bytes = new Uint8Array(hex.length / 2);
  for (let i = 0; i < bytes.length; i++) {
    bytes[i] = parseInt(hex.slice(i * 2, i * 2 + 2), 16);
  }
  return bytes;
}

/**
 * Generate random genesis salt (16 bytes)
 */
export function generateGenesisSalt(): Uint8Array {
  const salt = new Uint8Array(GENESIS_SALT_SIZE);
  crypto.getRandomValues(salt);
  return salt;
}
