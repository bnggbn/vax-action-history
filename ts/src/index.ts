/**
 * VAX JCS - TypeScript implementation
 *
 * JSON Canonicalization Scheme (JCS) for VAX action history
 */

export {
  marshal,
  canonicalizeJSON,
  canonicalizeValue,
  normalizeJSONNumber
} from './jcs';

export {
  computeSAI,
  computeGenesisSAI,
  verifyAction,
  toHex,
  fromHex,
  generateGenesisSalt,
  SAI_SIZE,
  GI_SIZE,
  GENESIS_SALT_SIZE,
  VaxError,
  InvalidInputError,
  InvalidPrevSAIError,
} from './vax';
