/**
 * VAX TypeScript SDK Tests
 */

import {
  computeSAI,
  computeGenesisSAI,
  verifyAction,
  toHex,
  fromHex,
  generateGenesisSalt,
  SAI_SIZE,
  GI_SIZE,
  GENESIS_SALT_SIZE,
  InvalidInputError,
  InvalidPrevSAIError,
} from './vax';

describe('VAX SDK', () => {
  // Test vectors (matching Go/C test suites)
  const testGenesisSalt = new Uint8Array([
    0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8,
    0xa9, 0xaa, 0xab, 0xac, 0xad, 0xae, 0xaf, 0xb0,
  ]);

  describe('computeGenesisSAI', () => {
    it('should produce 32-byte output', async () => {
      const actorID = 'user123:device456';
      const sai = await computeGenesisSAI(actorID, testGenesisSalt);
      expect(sai.length).toBe(SAI_SIZE);
    });

    it('should match known test vector', async () => {
      const actorID = 'user123:device456';
      const sai = await computeGenesisSAI(actorID, testGenesisSalt);

      // Expected from Go/C test suites
      const expected = 'afc50728cd79e805a8ae06875a1ddf78ca11b0d56ec300b160fb71f50ce658c3';
      expect(toHex(sai)).toBe(expected);
    });

    it('should throw on invalid genesis_salt length', async () => {
      await expect(computeGenesisSAI('test', new Uint8Array([0x01, 0x02])))
        .rejects.toThrow(InvalidInputError);
    });
  });

  describe('computeSAI', () => {
    it('should produce 32-byte output', async () => {
      const prevSAI = new Uint8Array(SAI_SIZE).fill(0x11);
      const sae = new TextEncoder().encode('{"action":"test","value":42}');

      const sai = await computeSAI(prevSAI, sae);
      expect(sai.length).toBe(SAI_SIZE);
    });

    it('should produce different output each time (random gi)', async () => {
      const prevSAI = new Uint8Array(SAI_SIZE).fill(0x00);
      const sae = new TextEncoder().encode('{"test":1}');

      const sai1 = await computeSAI(prevSAI, sae);
      const sai2 = await computeSAI(prevSAI, sae);

      // Since gi is random, same inputs should produce different outputs
      expect(toHex(sai1)).not.toBe(toHex(sai2));
    });

    it('should throw on invalid prevSAI length', async () => {
      await expect(computeSAI(new Uint8Array([0x01]), new Uint8Array([0x01])))
        .rejects.toThrow(InvalidInputError);
    });

    it('should throw on empty SAE', async () => {
      await expect(computeSAI(new Uint8Array(SAI_SIZE), new Uint8Array(0)))
        .rejects.toThrow(InvalidInputError);
    });
  });

  describe('verifyAction', () => {
    it('should pass for matching prevSAI', () => {
      const prevSAI = new Uint8Array(SAI_SIZE).fill(0xAA);
      const expectedPrevSAI = new Uint8Array(SAI_SIZE).fill(0xAA);

      expect(() => verifyAction(expectedPrevSAI, prevSAI)).not.toThrow();
    });

    it('should throw InvalidPrevSAIError for non-matching prevSAI', () => {
      const expectedPrevSAI = new Uint8Array(SAI_SIZE).fill(0xAA);
      const wrongPrevSAI = new Uint8Array(SAI_SIZE).fill(0xBB);

      expect(() => verifyAction(expectedPrevSAI, wrongPrevSAI))
        .toThrow(InvalidPrevSAIError);
    });

    it('should throw InvalidInputError for wrong expectedPrevSAI length', () => {
      expect(() => verifyAction(new Uint8Array([0x01]), new Uint8Array(SAI_SIZE)))
        .toThrow(InvalidInputError);
    });

    it('should throw InvalidInputError for wrong prevSAI length', () => {
      expect(() => verifyAction(new Uint8Array(SAI_SIZE), new Uint8Array([0x01])))
        .toThrow(InvalidInputError);
    });
  });

  describe('Chain Simulation', () => {
    it('should successfully chain multiple actions', async () => {
      const actorID = 'alice:laptop';
      const genesisSalt = new Uint8Array(GENESIS_SALT_SIZE).fill(0xAB);

      // Genesis
      const genesisSAI = await computeGenesisSAI(actorID, genesisSalt);
      expect(genesisSAI.length).toBe(SAI_SIZE);

      // Action 1
      const sae1 = new TextEncoder().encode('{"action":"create","id":1}');
      const sai1 = await computeSAI(genesisSAI, sae1);
      expect(sai1.length).toBe(SAI_SIZE);

      // Action 2
      const sae2 = new TextEncoder().encode('{"action":"update","id":1}');
      const sai2 = await computeSAI(sai1, sae2);
      expect(sai2.length).toBe(SAI_SIZE);

      // Verify chain properties
      expect(toHex(sai1)).not.toBe(toHex(sai2));
    });
  });

  describe('Utility Functions', () => {
    describe('toHex / fromHex', () => {
      it('should round-trip correctly', () => {
        const original = new Uint8Array([0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef]);
        const hex = toHex(original);
        const restored = fromHex(hex);

        expect(toHex(restored)).toBe(toHex(original));
      });

      it('should convert to lowercase hex', () => {
        const bytes = new Uint8Array([0xAB, 0xCD, 0xEF]);
        expect(toHex(bytes)).toBe('abcdef');
      });

      it('should throw on odd-length hex string', () => {
        expect(() => fromHex('abc')).toThrow(InvalidInputError);
      });
    });

    describe('generateGenesisSalt', () => {
      it('should produce 16-byte output', () => {
        const salt = generateGenesisSalt();
        expect(salt.length).toBe(GENESIS_SALT_SIZE);
      });

      it('should produce different values each time', () => {
        const salt1 = generateGenesisSalt();
        const salt2 = generateGenesisSalt();
        expect(toHex(salt1)).not.toBe(toHex(salt2));
      });
    });
  });
});
