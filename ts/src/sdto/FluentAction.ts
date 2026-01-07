import { FieldSpec } from "./FieldSpec";
import { buildSAE } from "../sae";

// FluentAction is the fluent builder for constructing validated actions
export class FluentAction {
  private actionType: string;
  private schema: Record<string, FieldSpec>;
  private data: Record<string, unknown> = {};
  private errors: Error[] = [];

  constructor(actionType: string, schema: Record<string, FieldSpec>) {
    this.actionType = actionType;
    this.schema = schema;
  }

  // Set validates and stores a field value
  set(key: string, value: unknown): this {
    const spec = this.schema[key];
    if (!spec) {
      this.errors.push(new Error(`unknown field: ${key}`));
      return this;
    }

    const err = validateValue(value, spec);
    if (err) {
      this.errors.push(new Error(`field ${key}: ${err.message}`));
      return this;
    }

    this.data[key] = value;
    return this;
  }

  // Finalize returns SAE bytes or throws aggregated errors
  // Matches Go's Finalize() which returns ([]byte, error)
  finalize(): Buffer {
    // Check for missing required fields (all schema fields are required)
    for (const key of Object.keys(this.schema)) {
      if (!(key in this.data)) {
        this.errors.push(new Error(`missing required field: ${key}`));
      }
    }

    if (this.errors.length > 0) {
      const messages = this.errors.map((e) => e.message).join("; ");
      throw new Error(messages);
    }

    // Build SAE using JCS canonicalizer (matches Go's sae.BuildSAE)
    return buildSAE(this.actionType, this.data);
  }
}

// Validation helper functions (moved outside class to match Go pattern)

function validateValue(value: unknown, spec: FieldSpec): Error | null {
  switch (spec.type) {
    case "string":
      return validateString(value, spec);
    case "number":
      return validateNumber(value, spec);
    case "sign":
      return validateSign(value, spec);
    default:
      return new Error(`unknown type "${spec.type}"`);
  }
}

function validateSign(value: unknown, spec: FieldSpec): Error | null {
  // Sign value must be string (type is defined in schema)
  if (typeof value !== "string") {
    return new Error("sign field expects string value");
  }

  // Can extend: validate format based on spec.enum[0] (hex/base64 etc)
  if (value.length === 0) {
    return new Error("sign value cannot be empty");
  }

  return null;
}

function validateString(value: unknown, spec: FieldSpec): Error | null {
  if (typeof value !== "string") {
    return new Error("expected string");
  }

  // enum check
  if (spec.enum && spec.enum.length > 0) {
    if (!spec.enum.includes(value)) {
      return new Error(`value "${value}" not in enum`);
    }
    return null;
  }

  // length boundary (numeric parsing)
  if (spec.min !== undefined) {
    const minLen = parseInt(spec.min, 10);
    if (!isNaN(minLen) && value.length < minLen) {
      return new Error(`string length ${value.length} < min ${minLen}`);
    }
  }
  if (spec.max !== undefined) {
    const maxLen = parseInt(spec.max, 10);
    if (!isNaN(maxLen) && value.length > maxLen) {
      return new Error(`string length ${value.length} > max ${maxLen}`);
    }
  }

  return null;
}

function validateNumber(value: unknown, spec: FieldSpec): Error | null {
  if (typeof value !== "number") {
    return new Error("expected number");
  }

  if (spec.min !== undefined) {
    const minVal = parseFloat(spec.min);
    if (!isNaN(minVal) && value < minVal) {
      return new Error("number < min");
    }
  }
  if (spec.max !== undefined) {
    const maxVal = parseFloat(spec.max);
    if (!isNaN(maxVal) && value > maxVal) {
      return new Error("number > max");
    }
  }

  return null;
}

// ValidateData validates a map against schema (for server-side verification)
// Matches Go's ValidateData function
export function validateData(
  data: Record<string, unknown>,
  schema: Record<string, FieldSpec>
): void {
  const errors: Error[] = [];

  // Check all required fields in schema exist
  for (const [key, spec] of Object.entries(schema)) {
    const value = data[key];
    if (value === undefined) {
      errors.push(new Error(`missing field: ${key}`));
      continue;
    }
    const err = validateValue(value, spec);
    if (err) {
      errors.push(new Error(`field ${key}: ${err.message}`));
    }
  }

  // Check no extra fields
  for (const key of Object.keys(data)) {
    if (!(key in schema)) {
      errors.push(new Error(`unknown field: ${key}`));
    }
  }

  if (errors.length > 0) {
    throw new Error(errors.map((e) => e.message).join("; "));
  }
}

// Factory function to match Go's NewAction
export function newAction(
  actionType: string,
  schema: Record<string, FieldSpec>
): FluentAction {
  return new FluentAction(actionType, schema);
}
