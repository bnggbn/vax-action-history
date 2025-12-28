/**
 * VAX JSON Canonicalization Scheme (JCS)
 */

export function canonicalizeValue(value: unknown): Buffer {
  const result: string[] = [];
  writeCanonicalValue(result, value);
  return Buffer.from(result.join(''), 'utf8');
}

export function canonicalizeJSON(input: Buffer | string): Buffer {
  const json = typeof input === 'string' ? input : input.toString('utf8');
  const value = JSON.parse(json);
  return canonicalizeValue(value);
}

export function marshal(value: unknown): Buffer {
  return canonicalizeValue(value);
}

// ======== Internal Implementation ========

function writeCanonicalValue(out: string[], value: unknown): void {
  if (value === null) {
    out.push('null');
  } else if (typeof value === 'boolean') {
    out.push(value ? 'true' : 'false');
  } else if (typeof value === 'string') {
    writeJSONString(out, value);
  } else if (typeof value === 'number') {
    out.push(formatNumber(value));
  } else if (Array.isArray(value)) {
    writeCanonicalArray(out, value);
  } else if (typeof value === 'object') {
    writeCanonicalObject(out, value as Record<string, unknown>);
  } else {
    throw new Error(`Unsupported type in canonical encoder: ${typeof value}`);
  }
}

function writeCanonicalObject(out: string[], obj: Record<string, unknown>): void {
  out.push('{');
  const keys = Object.keys(obj).sort();
  for (let i = 0; i < keys.length; i++) {
    if (i > 0) out.push(',');
    const key = keys[i];
    writeJSONString(out, key);
    out.push(':');
    writeCanonicalValue(out, obj[key]);
  }
  out.push('}');
}

function writeCanonicalArray(out: string[], arr: unknown[]): void {
  out.push('[');
  for (let i = 0; i < arr.length; i++) {
    if (i > 0) out.push(',');
    writeCanonicalValue(out, arr[i]);
  }
  out.push(']');
}

function writeJSONString(out: string[], str: string): void {
  out.push('"');

  for (let i = 0; i < str.length; i++) {
    const char = str[i];
    const code = str.charCodeAt(i);

    switch (char) {
      case '"':
        out.push('\\"');
        break;
      case '\\':
        out.push('\\\\');
        break;
      case '\b':
        out.push('\\b');
        break;
      case '\f':
        out.push('\\f');
        break;
      case '\n':
        out.push('\\n');
        break;
      case '\r':
        out.push('\\r');
        break;
      case '\t':
        out.push('\\t');
        break;
      default:
        if (code < 0x20) {
          // Control characters → \u00XX
          out.push('\\u00');
          out.push(hex2(code));
        } else if (code >= 0x20 && code <= 0x7E) {
          // Printable ASCII
          out.push(char);
        } else {
          // Non-ASCII → escape as UTF-16 code units

          // Handle surrogate pairs
          if (code >= 0xD800 && code <= 0xDBFF && i + 1 < str.length) {
            const low = str.charCodeAt(i + 1);
            if (low >= 0xDC00 && low <= 0xDFFF) {
              // Valid surrogate pair
              out.push('\\u');
              out.push(hex4(code));
              out.push('\\u');
              out.push(hex4(low));
              i++; // Skip the low surrogate
              break;
            }
          }

          // Single code unit
          out.push('\\u');
          out.push(hex4(code));
        }
    }
  }

  out.push('"');
}

function formatNumber(num: number): string {
  // Normalize -0 to 0
  if (num === 0 || Object.is(num, -0)) {
    return '0';
  }

  // Reject NaN explicitly
  if (isNaN(num)) {
    throw new Error('NaN is not allowed in VAX-JCS');
  }

  // Reject Infinity
  if (!isFinite(num)) {
    throw new Error('Infinity is not allowed in VAX-JCS');
  }

  return toPlainNumberString(num);
}

function hex2(byte: number): string {
  return byte.toString(16).padStart(2, '0');
}

function hex4(code: number): string {
  return code.toString(16).padStart(4, '0');
}

export function normalizeJSONNumber(raw: string): string {
  const decimalPattern = /^-?[0-9]+(\.[0-9]+)?$/;

  if (!decimalPattern.test(raw)) {
    throw new Error(`Non-decimal number not allowed: ${raw}`);
  }

  const isNegative = raw.startsWith('-');
  const unsigned = isNegative ? raw.slice(1) : raw;
  const hasDecimal = raw.includes('.');

  if (
    unsigned.length > 1 &&
    unsigned.startsWith('0') &&
    !(hasDecimal && unsigned.startsWith('0.'))
  ) {
    if (isNegative) {
      throw new Error(`Invalid leading zero: ${raw}`);
    }
    throw new Error(`Non-decimal number not allowed: ${raw}`);
  }

  const num = parseFloat(raw);

  if (!isFinite(num)) {
    throw new Error(`Invalid JSON number: ${raw}`);
  }

  if (num === 0) {
    return '0';
  }

  return toPlainNumberString(num);
}

function toPlainNumberString(num: number): string {
  const str = num.toString();
  const sciIndex = str.search(/[eE]/);

  if (sciIndex === -1) {
    return str;
  }

  const [mantissa, exponentPart] = str.toLowerCase().split('e');
  const exponent = parseInt(exponentPart, 10);

  // Reject extreme exponents to keep numbers within a manageable, predictable range
  if (Math.abs(exponent) > 12) {
    throw new Error(
      `Number too large/small for VAX-JCS (would use scientific notation): ${num}`
    );
  }

  const sign = mantissa.startsWith('-') ? '-' : '';
  const unsignedMantissa = mantissa.replace('-', '');
  const parts = unsignedMantissa.split('.');
  const intPart = parts[0];
  const fracPart = parts[1] ?? '';
  const digits = intPart + fracPart;

  const decimalIndex = intPart.length + exponent;
  let plain: string;

  if (decimalIndex <= 0) {
    plain = '0.' + '0'.repeat(-decimalIndex) + digits;
  } else if (decimalIndex >= digits.length) {
    plain = digits + '0'.repeat(decimalIndex - digits.length);
  } else {
    plain = `${digits.slice(0, decimalIndex)}.${digits.slice(decimalIndex)}`;
  }

  // Trim trailing zeros in fractional part
  if (plain.includes('.')) {
    plain = plain.replace(/\.?0+$/, '').replace(/\.$/, '');
  }

  return sign + plain;
}
