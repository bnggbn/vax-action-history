import { marshal } from './jcs';
import fc from 'fast-check';

describe('Property-based tests', () => {
  test('marshal then parse should be idempotent', () => {
    fc.assert(
      fc.property(
        fc.oneof(
          fc.constant(null),
          fc.boolean(),
          fc.integer(),
          fc.double({ noNaN: true, noDefaultInfinity: true }).filter(n => !n.toString().includes('e')),
          fc.string(),
          fc.array(fc.oneof(fc.constant(null), fc.boolean(), fc.integer(), fc.string())),
          fc.dictionary(fc.string(), fc.oneof(fc.constant(null), fc.boolean(), fc.integer(), fc.string()))
        ),
        (value) => {
          const canonical = marshal(value).toString();
          const parsed = JSON.parse(canonical);
          const canonical2 = marshal(parsed).toString();
          expect(canonical).toBe(canonical2);
        }
      ),
      { numRuns: 100 }
    );
  });

  test('valid Unicode object keys follow UTF-8 byte order', () => {
    fc.assert(
      fc.property(
        fc.dictionary(fc.fullUnicodeString(), fc.integer()),
        (obj) => {
          const canonical = marshal(obj).toString();
          const keysInText = [...canonical.matchAll(/"((?:[^"\\]|\\.)*)":/g)]
            .map(match => JSON.parse(`"${match[1]}"`));
          // Valid UTF-8 preserves scalar-value order. Use a scalar reference
          // independently of the production Buffer comparator.
          const expected = Object.keys(obj).sort((a, b) => {
            const left = Array.from(a, char => char.codePointAt(0)!);
            const right = Array.from(b, char => char.codePointAt(0)!);
            for (let i = 0; i < Math.min(left.length, right.length); i++) {
              if (left[i] !== right[i]) return left[i] - right[i];
            }
            return left.length - right.length;
          });
          expect(keysInText).toEqual(expected);
        }
      ),
      { numRuns: 200 }
    );
  });

  test('no whitespace in output', () => {
    fc.assert(
      fc.property(
        fc.oneof(
          fc.array(fc.integer()),
          fc.dictionary(fc.string(), fc.integer())
        ),
        (value) => {
          const canonical = marshal(value).toString();
          const withoutStrings = canonical.replace(/"(?:[^"\\]|\\.)*"/g, '""');
          expect(withoutStrings).not.toMatch(/\s/);
        }
      ),
      { numRuns: 100 }
    );
  });

  test('unicode characters are escaped', () => {
    fc.assert(
      fc.property(
        fc.unicode().filter(s => s.charCodeAt(0) > 0x7E),
        (str) => {
          const canonical = marshal(str).toString();
          // Should contain \u escapes
          expect(canonical).toMatch(/\\u[0-9a-f]{4}/i);
        }
      ),
      { numRuns: 50 }
    );
  });

  test('arrays preserve order', () => {
    fc.assert(
      fc.property(
        fc.array(fc.integer({ min: 0, max: 1000 }), { minLength: 2 }),
        (arr) => {
          const canonical = marshal(arr).toString();
          const parsed = JSON.parse(canonical);
          expect(parsed).toEqual(arr);
        }
      ),
      { numRuns: 100 }
    );
  });
});
