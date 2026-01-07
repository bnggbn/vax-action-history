import { marshal } from './jcs';
import testVectors from '../../../test-vectors.json';

interface TestVector {
  name: string;
  input: unknown;
  expected: string;
}

describe('Cross-language compatibility', () => {
  (testVectors as unknown as TestVector[]).forEach(({ name, input, expected }) => {
    test(name, () => {
      const result = marshal(input).toString();
      expect(result).toBe(expected);
    });
  });
});
