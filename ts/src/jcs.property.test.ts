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

  test('object keys always sorted', () => {
    fc.assert(
      fc.property(
        fc.dictionary(fc.string(), fc.oneof(fc.constant(null), fc.boolean(), fc.integer())),
        (obj) => {
          if (Object.keys(obj).length < 2) return; // Skip if too few keys

          const canonical = marshal(obj).toString();
          const keys = Object.keys(obj).sort();

          // Check keys appear in sorted order
          for (let i = 0; i < keys.length - 1; i++) {
            const key1 = keys[i];
            const key2 = keys[i + 1];
            const idx1 = canonical.indexOf(`"${key1}":`);
            const idx2 = canonical.indexOf(`"${key2}":`);
            expect(idx1).toBeLessThan(idx2);
          }
        }
      ),
      { numRuns: 100 }
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
          expect(canonical).not.toMatch(/\s/);
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
