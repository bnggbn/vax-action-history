import { marshal, canonicalizeJSON, canonicalizeValue, normalizeJSONNumber } from './jcs';

describe('marshal', () => {
  describe('basic types', () => {
    test('null', () => {
      expect(marshal(null).toString()).toBe('null');
    });

    test('bool true', () => {
      expect(marshal(true).toString()).toBe('true');
    });

    test('bool false', () => {
      expect(marshal(false).toString()).toBe('false');
    });

    test('int', () => {
      expect(marshal(123).toString()).toBe('123');
    });

    test('negative int', () => {
      expect(marshal(-456).toString()).toBe('-456');
    });

    test('float64', () => {
      expect(marshal(123.456).toString()).toBe('123.456');
    });

    test('string', () => {
      expect(marshal('hello').toString()).toBe('"hello"');
    });
  });

  describe('arrays', () => {
    test('empty array', () => {
      expect(marshal([]).toString()).toBe('[]');
    });

    test('int array', () => {
      expect(marshal([1, 2, 3]).toString()).toBe('[1,2,3]');
    });

    test('string array', () => {
      expect(marshal(['a', 'b', 'c']).toString()).toBe('["a","b","c"]');
    });

    test('mixed array', () => {
      expect(marshal([1, 'two', true, null]).toString()).toBe('[1,"two",true,null]');
    });
  });

  describe('objects', () => {
    test('empty object', () => {
      expect(marshal({}).toString()).toBe('{}');
    });

    test('simple object', () => {
      const obj = { name: 'Alice', age: 30 };
      expect(marshal(obj).toString()).toBe('{"age":30,"name":"Alice"}');
    });

    test('map with unsorted keys', () => {
      const obj = { z: 1, a: 2, m: 3 };
      expect(marshal(obj).toString()).toBe('{"a":2,"m":3,"z":1}');
    });
  });

  describe('nested structures', () => {
    test('map with nested array', () => {
      const obj = { items: [3, 1, 2], name: 'test' };
      expect(marshal(obj).toString()).toBe('{"items":[3,1,2],"name":"test"}');
    });

    test('array with nested map', () => {
      const arr = [
        { z: 1, a: 2 },
        { y: 3, b: 4 }
      ];
      expect(marshal(arr).toString()).toBe('[{"a":2,"z":1},{"b":4,"y":3}]');
    });
  });

  describe('special characters', () => {
    test('string with quotes', () => {
      expect(marshal('hello"world').toString()).toBe('"hello\\"world"');
    });

    test('string with backslash', () => {
      expect(marshal('path\\to\\file').toString()).toBe('"path\\\\to\\\\file"');
    });

    test('string with newline', () => {
      expect(marshal('line1\nline2').toString()).toBe('"line1\\nline2"');
    });

    test('string with unicode', () => {
      expect(marshal('你好').toString()).toBe('"\\u4f60\\u597d"');
    });
  });

  describe('number normalization', () => {
    test('negative zero float', () => {
      expect(marshal(-0.0).toString()).toBe('0');
    });

    test('Object.is -0', () => {
      const negZero = -0;
      expect(Object.is(negZero, -0)).toBe(true);
      expect(marshal(negZero).toString()).toBe('0');
    });
  });

  describe('complex examples', () => {
    test('realistic user object', () => {
      const user = {
        id: 12345,
        username: 'alice',
        email: 'alice@example.com',
        tags: ['admin', 'developer'],
        active: true
      };
      expect(marshal(user).toString()).toBe(
        '{"active":true,"email":"alice@example.com","id":12345,"tags":["admin","developer"],"username":"alice"}'
      );
    });
  });

  describe('error handling', () => {
    test('NaN throws error', () => {
      expect(() => marshal(NaN)).toThrow('NaN is not allowed in VAX-JCS');
    });

    test('Infinity throws error', () => {
      expect(() => marshal(Infinity)).toThrow('Infinity is not allowed in VAX-JCS');
    });

    test('-Infinity throws error', () => {
      expect(() => marshal(-Infinity)).toThrow('Infinity is not allowed in VAX-JCS');
    });

    test('Very large number (becomes Infinity) throws error', () => {
      expect(() => marshal(1e500)).toThrow('Infinity is not allowed in VAX-JCS');
    });

    test('Very large number (scientific notation) throws error', () => {
      expect(() => marshal(1e100)).toThrow('Number too large/small for VAX-JCS');
    });

    test('Very small number (scientific notation) throws error', () => {
      expect(() => marshal(1e-100)).toThrow('Number too large/small for VAX-JCS');
    });
  });
});

describe('normalizeJSONNumber', () => {
  describe('valid integers', () => {
    test('positive integer', () => {
      expect(normalizeJSONNumber('123')).toBe('123');
    });

    test('negative integer', () => {
      expect(normalizeJSONNumber('-456')).toBe('-456');
    });

    test('zero', () => {
      expect(normalizeJSONNumber('0')).toBe('0');
    });

    test('negative zero integer', () => {
      expect(normalizeJSONNumber('-0')).toBe('0');
    });

    test('large positive integer', () => {
      expect(normalizeJSONNumber('9007199254740991')).toBe('9007199254740991');
    });

    test('large negative integer', () => {
      expect(normalizeJSONNumber('-9007199254740991')).toBe('-9007199254740991');
    });
  });

  describe('valid decimals', () => {
    test('positive decimal', () => {
      expect(normalizeJSONNumber('123.456')).toBe('123.456');
    });

    test('negative decimal', () => {
      expect(normalizeJSONNumber('-123.456')).toBe('-123.456');
    });

    test('decimal starting with zero', () => {
      expect(normalizeJSONNumber('0.5')).toBe('0.5');
    });

    test('negative decimal starting with zero', () => {
      expect(normalizeJSONNumber('-0.5')).toBe('-0.5');
    });

    test('small decimal', () => {
      expect(normalizeJSONNumber('0.0001')).toBe('0.0001');
    });

    test('negative zero decimal', () => {
      expect(normalizeJSONNumber('-0.0')).toBe('0');
    });
  });

  describe('reject scientific notation', () => {
    test('scientific notation lowercase', () => {
      expect(() => normalizeJSONNumber('1e10')).toThrow('Non-decimal number not allowed');
    });

    test('scientific notation uppercase', () => {
      expect(() => normalizeJSONNumber('1E10')).toThrow('Non-decimal number not allowed');
    });

    test('scientific notation negative exponent', () => {
      expect(() => normalizeJSONNumber('1.5e-3')).toThrow('Non-decimal number not allowed');
    });

    test('scientific notation positive exponent', () => {
      expect(() => normalizeJSONNumber('2.5e+2')).toThrow('Non-decimal number not allowed');
    });
  });

  describe('reject leading zeros', () => {
    test('positive leading zero', () => {
      expect(() => normalizeJSONNumber('01')).toThrow('Non-decimal number not allowed');
    });

    test('negative leading zero', () => {
      expect(() => normalizeJSONNumber('-01')).toThrow('Invalid leading zero');
    });

    test('multiple leading zeros', () => {
      expect(() => normalizeJSONNumber('00')).toThrow('Non-decimal number not allowed');
    });

    test('negative multiple leading zeros', () => {
      expect(() => normalizeJSONNumber('-00')).toThrow('Invalid leading zero');
    });

    test('leading zero with decimal', () => {
      expect(() => normalizeJSONNumber('01.5')).toThrow('Non-decimal number not allowed');
    });
  });

  describe('reject invalid formats', () => {
    test('trailing dot', () => {
      expect(() => normalizeJSONNumber('1.')).toThrow('Non-decimal number not allowed');
    });

    test('leading dot', () => {
      expect(() => normalizeJSONNumber('.5')).toThrow('Non-decimal number not allowed');
    });

    test('multiple dots', () => {
      expect(() => normalizeJSONNumber('1.2.3')).toThrow('Non-decimal number not allowed');
    });

    test('empty string', () => {
      expect(() => normalizeJSONNumber('')).toThrow('Non-decimal number not allowed');
    });

    test('only minus sign', () => {
      expect(() => normalizeJSONNumber('-')).toThrow('Non-decimal number not allowed');
    });

    test('letters', () => {
      expect(() => normalizeJSONNumber('abc')).toThrow('Non-decimal number not allowed');
    });

    test('hex notation', () => {
      expect(() => normalizeJSONNumber('0x1F')).toThrow('Non-decimal number not allowed');
    });
  });

  describe('edge cases', () => {
    test('very small positive decimal', () => {
      expect(normalizeJSONNumber('0.000000000001')).toBe('0.000000000001');
    });

    test('very small negative decimal', () => {
      expect(normalizeJSONNumber('-0.000000000001')).toBe('-0.000000000001');
    });
  });
});

describe('canonicalizeJSON', () => {
  describe('basic types', () => {
    test('null', () => {
      expect(canonicalizeJSON('null').toString()).toBe('null');
    });

    test('true', () => {
      expect(canonicalizeJSON('true').toString()).toBe('true');
    });

    test('false', () => {
      expect(canonicalizeJSON('false').toString()).toBe('false');
    });

    test('integer', () => {
      expect(canonicalizeJSON('123').toString()).toBe('123');
    });

    test('negative integer', () => {
      expect(canonicalizeJSON('-456').toString()).toBe('-456');
    });

    test('decimal', () => {
      expect(canonicalizeJSON('123.456').toString()).toBe('123.456');
    });

    test('string', () => {
      expect(canonicalizeJSON('"hello"').toString()).toBe('"hello"');
    });
  });

  describe('empty structures', () => {
    test('empty object', () => {
      expect(canonicalizeJSON('{}').toString()).toBe('{}');
    });

    test('empty array', () => {
      expect(canonicalizeJSON('[]').toString()).toBe('[]');
    });
  });

  describe('key sorting', () => {
    test('object keys sorted', () => {
      expect(canonicalizeJSON('{"z":1,"a":2,"m":3}').toString()).toBe('{"a":2,"m":3,"z":1}');
    });

    test('nested object keys sorted', () => {
      expect(canonicalizeJSON('{"outer":{"z":1,"a":2}}').toString()).toBe('{"outer":{"a":2,"z":1}}');
    });
  });

  describe('whitespace removal', () => {
    test('remove whitespace in object', () => {
      expect(canonicalizeJSON('{ "a" : 1 , "b" : 2 }').toString()).toBe('{"a":1,"b":2}');
    });

    test('remove whitespace in array', () => {
      expect(canonicalizeJSON('[ 1 , 2 , 3 ]').toString()).toBe('[1,2,3]');
    });

    test('remove newlines and tabs', () => {
      const input = `{
        "a": 1,
        "b": 2
      }`;
      expect(canonicalizeJSON(input).toString()).toBe('{"a":1,"b":2}');
    });
  });

  describe('string escaping', () => {
    test('escape quote', () => {
      expect(canonicalizeJSON('"hello\\"world"').toString()).toBe('"hello\\"world"');
    });

    test('escape backslash', () => {
      expect(canonicalizeJSON('"path\\\\to\\\\file"').toString()).toBe('"path\\\\to\\\\file"');
    });

    test('escape newline', () => {
      expect(canonicalizeJSON('"line1\\nline2"').toString()).toBe('"line1\\nline2"');
    });

    test('escape tab', () => {
      expect(canonicalizeJSON('"col1\\tcol2"').toString()).toBe('"col1\\tcol2"');
    });
  });

  describe('unicode handling', () => {
    test('unicode escape', () => {
      expect(canonicalizeJSON('"你好"').toString()).toBe('"\\u4f60\\u597d"');
    });

    test('emoji', () => {
      expect(canonicalizeJSON('"😀"').toString()).toBe('"\\ud83d\\ude00"');
    });

    test('mixed ascii and unicode', () => {
      expect(canonicalizeJSON('"Hello世界"').toString()).toBe('"Hello\\u4e16\\u754c"');
    });
  });

  describe('number normalization', () => {
    test('negative zero', () => {
      expect(canonicalizeJSON('-0').toString()).toBe('0');
    });

    test('negative zero decimal', () => {
      expect(canonicalizeJSON('-0.0').toString()).toBe('0');
    });
  });

  describe('nested structures', () => {
    test('nested object and array', () => {
      expect(canonicalizeJSON('{"a":{"b":[1,2,3]},"c":null}').toString()).toBe('{"a":{"b":[1,2,3]},"c":null}');
    });

    test('complex nested', () => {
      expect(canonicalizeJSON('{"z":[{"y":2,"x":1}],"a":{"c":3,"b":4}}').toString())
        .toBe('{"a":{"b":4,"c":3},"z":[{"x":1,"y":2}]}');
    });
  });

  describe('error handling', () => {
    test('invalid json', () => {
      expect(() => canonicalizeJSON('{invalid}')).toThrow();
    });

    test('trailing comma', () => {
      expect(() => canonicalizeJSON('{"a":1,}')).toThrow();
    });
  });
});

describe('canonicalizeValue', () => {
  test('object', () => {
    const input = { b: 2, a: 1 };
    expect(canonicalizeValue(input).toString()).toBe('{"a":1,"b":2}');
  });
});
