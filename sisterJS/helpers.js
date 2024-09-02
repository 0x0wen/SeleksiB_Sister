function parseJSON(jsonString) {
  let index = 0;
  const length = jsonString.length;

  function parseValue() {
    skipWhitespace();
    if (index >= length) throw new SyntaxError('Unexpected end of input');
    const char = jsonString[index];

    if (char === '"') return parseString();
    if (char === '{') return parseObject();
    if (char === '[') return parseArray();
    if (char === 't' || char === 'f') return parseBoolean();
    if (char === 'n') return parseNull();
    if (char >= '0' && char <= '9' || char === '-') return parseNumber();
    
    throw new SyntaxError('Unexpected character');
  }

  function skipWhitespace() {
    while (index < length && /\s/.test(jsonString[index])) index++;
  }

  function parseString() {
    index++; // Skip opening quote
    let result = '';
    while (index < length && jsonString[index] !== '"') {
      if (jsonString[index] === '\\') {
        index++;
        if (index < length) {
          const escapeChar = jsonString[index];
          if (escapeChar === 'n') result += '\n';
          else if (escapeChar === 't') result += '\t';
          else if (escapeChar === 'r') result += '\r';
          else if (escapeChar === 'b') result += '\b';
          else if (escapeChar === 'f') result += '\f';
          else if (escapeChar === '"') result += '"';
          else if (escapeChar === '\\') result += '\\';
          else throw new SyntaxError('Invalid escape character');
        }
      } else {
        result += jsonString[index];
      }
      index++;
    }
    index++; // Skip closing quote
    return result;
  }

  function parseObject() {
    index++; // Skip opening brace
    const result = {};
    skipWhitespace();
    if (jsonString[index] === '}') {
      index++; // Skip closing brace
      return result;
    }
    while (true) {
      skipWhitespace();
      const key = parseString();
      skipWhitespace();
      if (jsonString[index] !== ':') throw new SyntaxError('Expected colon');
      index++; // Skip colon
      const value = parseValue();
      result[key] = value;
      skipWhitespace();
      if (jsonString[index] === '}') {
        index++;
        break;
      }
      if (jsonString[index] !== ',') throw new SyntaxError('Expected comma');
      index++; // Skip comma
    }
    return result;
  }

  function parseArray() {
    index++; // Skip opening bracket
    const result = [];
    skipWhitespace();
    if (jsonString[index] === ']') {
      index++; // Skip closing bracket
      return result;
    }
    while (true) {
      const value = parseValue();
      result.push(value);
      skipWhitespace();
      if (jsonString[index] === ']') {
        index++;
        break;
      }
      if (jsonString[index] !== ',') throw new SyntaxError('Expected comma');
      index++; // Skip comma
    }
    return result;
  }

  function parseBoolean() {
    const str = jsonString.slice(index, index + 4);
    if (str === 'true') {
      index += 4;
      return true;
    }
    const strFalse = jsonString.slice(index, index + 5);
    if (strFalse === 'false') {
      index += 5;
      return false;
    }
    throw new SyntaxError('Invalid boolean');
  }

  function parseNull() {
    const str = jsonString.slice(index, index + 4);
    if (str === 'null') {
      index += 4;
      return null;
    }
    throw new SyntaxError('Invalid null');
  }

  function parseNumber() {
    let start = index;
    if (jsonString[index] === '-') index++;
    while (index < length && /\d/.test(jsonString[index])) index++;
    if (jsonString[index] === '.') index++;
    while (index < length && /\d/.test(jsonString[index])) index++;
    const exponent = jsonString.slice(index).match(/^([eE][+-]?\d+)?/);
    if (exponent) index += exponent[0].length;
    const num = parseFloat(jsonString.slice(start, index));
    if (isNaN(num)) throw new SyntaxError('Invalid number');
    return num;
  }

  return parseValue();
}

// JSON stringifying function
function stringifyJSON(value) {
  if (typeof value === 'string') return `"${value.replace(/\\/g, '\\\\').replace(/"/g, '\\"')}"`;
  if (typeof value === 'number') return isFinite(value) ? Number(value).toString() : 'null'; // handle NaN and Infinity
  if (typeof value === 'boolean') return String(value);
  if (value === null) return 'null';
  if (Array.isArray(value)) {
    return `[${value.map(stringifyJSON).join(',')}]`;
  }
  if (typeof value === 'object') {
    const entries = Object.entries(value).map(([key, val]) => {
      if (typeof key !== 'string') throw new TypeError('Keys must be strings');
      return `"${key}":${stringifyJSON(val)}`;
    });
    return `{${entries.join(',')}}`;
  }
  throw new TypeError('Unsupported data type');
}


module.exports = { parseJSON, stringifyJSON };