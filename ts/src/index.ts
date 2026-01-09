/**
 * VAX TypeScript SDK
 *
 * Git-like tamper-evident action history with deterministic output.
 */

// Core VAX functions
export {
  computeSAI,
  computeGenesisSAI,
  verifyAction,
  verifyPrevSAI,
  toHex,
  fromHex,
  generateGenesisSalt,
  SAI_SIZE,
  GENESIS_SALT_SIZE,
  VaxError,
  InvalidInputError,
  InvalidPrevSAIError,
  SAIMismatchError,
} from './vax';

// JCS - JSON Canonicalization Scheme
export {
  marshal,
  canonicalizeJSON,
  canonicalizeValue,
  normalizeJSONNumber,
} from './jcs';

// SAE - Semantic Action Envelope
export {
  Envelope,
  buildSAE,
} from './sae';

// SDTO - Schema-Driven Type Objects
export {
  FieldSpec,
  parseSchema,
  FluentAction,
  newAction,
  validateData,
  SchemaBuilder,
  newSchemaBuilder,
} from './sdto';
