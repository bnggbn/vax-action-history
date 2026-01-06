import { FieldSpec } from "./FieldSpec";

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

    const err = this.validateValue(value, spec);
    if (err) {
      this.errors.push(new Error(`field ${key}: ${err.message}`));
      return this;
    }

    this.data[key] = value;
    return this;
  }

  // Finalize returns the action data or throws aggregated errors
  finalize(): { actionType: string; data: Record<string, unknown> } {
    if (this.errors.length > 0) {
      const messages = this.errors.map((e) => e.message).join("; ");
      throw new Error(messages);
    }
    return {
      actionType: this.actionType,
      data: this.data,
    };
  }

  private validateValue(value: unknown, spec: FieldSpec): Error | null {
    switch (spec.type) {
      case "string":
        return this.validateString(value, spec);
      case "number":
        return this.validateNumber(value, spec);
      default:
        return new Error(`unknown type "${spec.type}"`);
    }
  }

  private validateString(value: unknown, spec: FieldSpec): Error | null {
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

  private validateNumber(value: unknown, spec: FieldSpec): Error | null {
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
}

// Factory function to match Go's NewAction
export function newAction(
  actionType: string,
  schema: Record<string, FieldSpec>
): FluentAction {
  return new FluentAction(actionType, schema);
}
