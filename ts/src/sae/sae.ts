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
