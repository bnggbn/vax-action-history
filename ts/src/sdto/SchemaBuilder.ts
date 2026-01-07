import { FieldSpec } from "./FieldSpec";

// Supported sign types
export const SupportedSignTypes = ["ed25519", "rsa", "ecdsa"] as const;

// SchemaBuilder is the fluent builder for defining validation schemas
export class SchemaBuilder {
  private actions: Record<string, FieldSpec> = {};

  // SetActionStringLength sets string length constraints
  setActionStringLength(action: string, min: string, max: string): this {
    this.actions[action] = {
      type: "string",
      min,
      max,
    };
    return this;
  }

  // SetActionNumberRange sets number range constraints
  setActionNumberRange(action: string, min: string, max: string): this {
    this.actions[action] = {
      type: "number",
      min,
      max,
    };
    return this;
  }

  // SetActionEnum sets enum constraints
  setActionEnum(action: string, values: string[]): this {
    this.actions[action] = {
      type: "string",
      enum: values,
    };
    return this;
  }

  // SetActionSign sets sign field with specified sign algorithm type
  setActionSign(action: string, signType: string): this {
    this.actions[action] = {
      type: "sign",
      enum: [signType],
    };
    return this;
  }

  // SetActionSignMulti sets sign field allowing multiple sign types
  setActionSignMulti(action: string, signTypes: string[]): this {
    this.actions[action] = {
      type: "sign",
      enum: signTypes,
    };
    return this;
  }

  // BuildSchema returns the FieldSpec map for constructor use
  buildSchema(): Record<string, FieldSpec> {
    return { ...this.actions };
  }

  // Build returns JSON-friendly format for cross-service transport
  build(): Record<string, unknown> {
    const props: Record<string, unknown> = {};

    for (const [name, spec] of Object.entries(this.actions)) {
      const m: Record<string, unknown> = { type: spec.type };
      if (spec.min !== undefined) m["min"] = spec.min;
      if (spec.max !== undefined) m["max"] = spec.max;
      if (spec.enum && spec.enum.length > 0) m["enum"] = spec.enum;
      props[name] = m;
    }

    return {
      type: "object",
      properties: props,
    };
  }
}

// Factory function to match Go's NewSchemaBuilder
export function newSchemaBuilder(): SchemaBuilder {
  return new SchemaBuilder();
}
