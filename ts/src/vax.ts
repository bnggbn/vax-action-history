/**
 * VAX TypeScript SDK
 *
 * Git-like tamper-evident action history with deterministic output.
 */

// ============================================================================
// Constants
// ============================================================================

export const SAI_SIZE = 32;
export const GI_SIZE = 32;
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

// ============================================================================
// Internal Helpers
// ============================================================================

/**
 * Compute gi = random 32 bytes (256-bit)
 * Uses Web Crypto API for secure random generation.
 */
function computeGI(): Uint8Array {
  const gi = new Uint8Array(GI_SIZE);
  crypto.getRandomValues(gi);
  return gi;
}

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
 * Compute SAI_n = SHA256("VAX-SAI" || prevSAI || SHA256(SAE) || gi)
 *
 * gi is generated internally using random bytes.
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

  // Generate random gi
  const gi = computeGI();

  // message = "VAX-SAI" || prevSAI || saeHash || gi
  const label = stringToBytes('VAX-SAI');
  const message = concat(label, prevSAI, saeHash, gi);

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
 * Verify an action submission (crypto only, no JSON validation)
 *
 * This verifies that prevSAI matches the expected value.
 *
 * @param expectedPrevSAI - Expected previous SAI (32 bytes)
 * @param prevSAI - Submitted previous SAI (32 bytes)
 * @throws InvalidInputError if inputs are invalid
 * @throws InvalidPrevSAIError if prevSAI doesn't match
 */
export function verifyAction(
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
