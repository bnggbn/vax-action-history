/**
 * VAX Semantic Action Envelope (SAE)
 *
 * SAE is the core data structure for action history.
 */

import { marshal } from '../jcs';

export interface Envelope {
  action_type: string;
  timestamp: number;
  sdto: Record<string, unknown>;
  signature?: Uint8Array;
}

/**
 * BuildSAE builds a Semantic Action Envelope using the project's JCS canonicalizer.
 *
 * IMPORTANT: We do NOT use JSON.stringify()
 * We MUST ONLY use our own JCS canonicalizer.
 *
 * @param actionType - The type of action
 * @param sdto - Schema-Driven Type Objects data
 * @returns Buffer - JCS-canonicalized SAE bytes
 */
export function buildSAE(actionType: string, sdto: Record<string, unknown>): Buffer {
  const env: Envelope = {
    action_type: actionType,
    timestamp: Date.now(),
    sdto: sdto,
    // signature is omitted (undefined) for unsigned SAE
  };

  // IMPORTANT: Use JCS canonicalizer, not JSON.stringify
  return marshal(env);
}

/**
 * Sign an envelope with Ed25519 private key
 * Note: Ed25519 signing requires @noble/ed25519 or similar library in Node.js
 * This is a placeholder that will work when Web Crypto Ed25519 is available
 *
 * @param envelope - The envelope to sign
 * @param privateKey - Ed25519 private key (64 bytes)
 */
export async function signEnvelope(
  envelope: Envelope,
  privateKey: Uint8Array
): Promise<Envelope> {
  if (privateKey.length !== 64) {
    throw new Error('invalid Ed25519 private key');
  }

  // Get canonical bytes for signing (without signature field)
  const { signature: _, ...unsignedEnvelope } = envelope;
  const canonical = marshal(unsignedEnvelope);

  // Use Web Crypto API for Ed25519 signing (Node.js 18+ with experimental flag)
  // For production, consider using @noble/ed25519 library
  try {
    const key = await globalThis.crypto.subtle.importKey(
      'raw',
      privateKey.slice(0, 32), // seed portion
      { name: 'Ed25519' },
      false,
      ['sign']
    );

    const signatureBuffer = await globalThis.crypto.subtle.sign('Ed25519', key, canonical);
    envelope.signature = new Uint8Array(signatureBuffer);
  } catch {
    // Fallback: Ed25519 not supported in this environment
    // In production, use @noble/ed25519 or similar
    throw new Error('Ed25519 signing not supported in this environment. Consider using @noble/ed25519.');
  }

  return envelope;
}

/**
 * Generate an Ed25519 key pair
 * Note: Requires Web Crypto Ed25519 support (Node.js 18+ or modern browsers)
 * For broader compatibility, use @noble/ed25519 library
 *
 * @returns Promise<{ publicKey: Uint8Array, privateKey: Uint8Array }>
 */
export async function generateKeyPair(): Promise<{
  publicKey: Uint8Array;
  privateKey: Uint8Array;
}> {
  try {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const keyPair = await globalThis.crypto.subtle.generateKey(
      { name: 'Ed25519' } as any,
      true,
      ['sign', 'verify']
    );

    // Type assertion for key pair
    const kp = keyPair as { publicKey: unknown; privateKey: unknown };

    const publicKey = new Uint8Array(
      await globalThis.crypto.subtle.exportKey('raw', kp.publicKey as Parameters<typeof globalThis.crypto.subtle.exportKey>[1])
    );
    const privateKey = new Uint8Array(
      await globalThis.crypto.subtle.exportKey('pkcs8', kp.privateKey as Parameters<typeof globalThis.crypto.subtle.exportKey>[1])
    );

    return { publicKey, privateKey };
  } catch {
    throw new Error('Ed25519 key generation not supported in this environment. Consider using @noble/ed25519.');
  }
}
