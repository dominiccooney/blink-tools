module csv;

import std.encoding;
import std.stdio;

struct CharIterator {
  this(string s) {
    this.s = s;
  }

  dchar next() {
    if (bufferFull) {
      bufferFull = false;
      return buffer;
    }
    return std.encoding.decode(s);
  }

  dchar peek() {
    if (bufferFull) {
      return buffer;
    }
    buffer = next();
    bufferFull = true;
    return buffer;
  }

  bool atEnd() {
    return !s.length && !bufferFull;
  }

  dchar buffer;
  bool bufferFull;
  string s;
}

unittest {
  auto c = CharIterator("fold");
  assert(!c.atEnd());
  assert('f' == c.next());
  assert('o' == c.peek());
  assert('o' == c.next());
  assert('l' == c.next());
  assert(!c.atEnd());
  assert('d' == c.peek());
  assert(!c.atEnd());
  assert('d' == c.next());
  assert(c.atEnd());
}

struct CSVParser {
  this(File file) {
    this.file = file;
  }

  this(string[] lines) {
    this.rawLines = lines;
  }

  int opApply(int delegate(ref string[]) dg) {
    foreach (line; lines()) {
      string[] fields = parseLine(line);
      if (fields == null) {
        return 0;
      }
      if (int result = dg(fields)) {
        return result;
      }
    }
    return 0;
  }

  string[][] toArray() {
    string[][] result = null;
    foreach (fields; this) {
      result ~= fields;
    }
    return result;
  }

private:
  int delegate(int delegate(ref string)) lines() {
    if (rawLines) {
      return (dg) {
        foreach (line; rawLines) {
          if (int result = dg(line)) {
            return result;
          }
        }
        return 0;
      };
    } else {
      return (dg) {
        foreach (string line; std.stdio.lines(file)) {
          if (int result = dg(line)) {
            return result;
          }
        }
        return 0;
      };
    }
  }

  string[] parseLine(string line) {
    auto chars = CharIterator(line);
    return parse(chars, null);
  }

  string[] parse(ref CharIterator chars, string[] fields) {
    if (chars.atEnd()) {
      return fields;
    }

    return '"' == chars.peek()
        ? parseQuoted(chars, fields)
        : parseUnquoted(chars, fields);
  }

  string[] parseUnquoted(ref CharIterator chars, string[] fields) {
    auto field = "";
    while (!chars.atEnd()) {
      dchar ch = chars.next();
      if (',' == ch) {
        break;
      }
      field ~= ch;
    }
    return parse(chars, fields ~ field);
  }

  string[] parseQuoted(ref CharIterator chars, string[] fields) {
    chars.next(); // chomp "
    auto field = "";
    for (;;) {
      if (chars.atEnd()) {
        throw new Exception("unterminated quoted field");
      }

      if ('"' == chars.peek()) {
        chars.next();
        dchar ch = chars.atEnd() ? ',' : chars.next();
        if ('"' == ch) {
          field ~= ch;
        } else if (',' == ch) {
          break;
        } else {
          throw new Exception(
              std.string.format("unexpected character '%s' after '\"'", ch));
        }
      } else {
        field ~= chars.next();
      }
    }

    return parse(chars, fields ~ field);
  }

  File file;
  string[] rawLines;
}

unittest {
  assert([["foo"], ["bar"]] == CSVParser(["foo", "bar"]).toArray());
  assert(
      [["Element", "Weight"],
       ["Hydrogen", "1"]] ==
      CSVParser(["Element,Weight", "Hydrogen,1"]).toArray());
  assert(
      [["Person", "Weight"],
       ["W. Churchill", "Important"]] ==
      CSVParser(["Person,Weight", "\"W. Churchill\",Important"]).toArray());
  assert(
      [["Person", "Quote"],
       ["F.D. Roosevelt",
        "\"The only thing we have to fear is fear itself.\" (1933)"]] ==
      CSVParser(["Person,Quote",
                 "F.D. Roosevelt,\"\"\"The only thing we have to fear is "
                 "fear itself.\"\" (1933)\""]).toArray());
}
