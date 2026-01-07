// FieldSpec defines the validation rules for a field
export interface FieldSpec {
  type: "string" | "number" | "sign";
  min?: string;
  max?: string;
  enum?: string[];
}

// ParseSchema converts Record<string, unknown> to Record<string, FieldSpec>
// Used for cross-service deserialization
export function parseSchema(raw: Record<string, unknown>): Record<string, FieldSpec> {
  const result: Record<string, FieldSpec> = {};

  for (const [key, val] of Object.entries(raw)) {
    if (typeof val !== "object" || val === null) {
      continue;
    }

    const m = val as Record<string, unknown>;
    const spec: FieldSpec = {
      type: (m["type"] as "string" | "number" | "sign") ?? "string",
    };

    if (typeof m["min"] === "string") {
      spec.min = m["min"];
    }
    if (typeof m["max"] === "string") {
      spec.max = m["max"];
    }
    if (Array.isArray(m["enum"])) {
      spec.enum = m["enum"].filter((e): e is string => typeof e === "string");
    }

    result[key] = spec;
  }

  return result;
}
