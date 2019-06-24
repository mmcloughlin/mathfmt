package main

import (
	"bytes"
	"errors"
	"unicode"
)

// super is the replacement map for superscript characters.
var super = map[rune]rune{}

func init() {
	for _, char := range chars {
		if char.Super != None {
			super[char.Char] = char.Super
		}
	}
}

// Format processes the source code in b.
func Format(b []byte) ([]byte, error) {
	var buf bytes.Buffer

	for len(b) > 0 {
		switch {
		// Start of a comment.
		case prefix(b, "//"):
			rest, err := comment(&buf, b)
			if err != nil {
				return nil, err
			}
			b = rest
		default:
			buf.WriteByte(b[0])
			b = b[1:]
		}
	}

	return buf.Bytes(), nil
}

// comment processes a single line comment.
func comment(w *bytes.Buffer, b []byte) ([]byte, error) {
	for len(b) > 0 {
		// Stop at new line.
		if b[0] == '\n' {
			w.WriteByte(b[0])
			b = b[1:]
			return b, nil
		}

		// Look for a replacable symbol.
		for symbol, r := range symbols {
			if prefix(b, symbol) {
				w.WriteRune(r)
				b = b[len(symbol):]
				break
			}
		}

		// Is this a recognized symbol?
		// Is this the start of an exponent?
		if prefix(b, "2^") {
			rest, err := exp(w, b)
			if err != nil {
				return nil, err
			}
			b = rest
			continue
		}

		// Otherwise consume a byte.
		w.WriteByte(b[0])
		b = b[1:]
	}
	return b, nil
}

// Exponentiation represents an exponentiation of the form base^e.
type Exponentiation struct {
	Base     []byte
	Exponent []byte
	Raw      []byte
}

// exp processes an exponentiation.
func exp(w *bytes.Buffer, b []byte) ([]byte, error) {
	e, rest, err := parseexp(b)
	if err != nil {
		return nil, err
	}

	// Is the exponent replaceable with superscripts? If not write out unchanged and return.
	exponent := bytes.Runes(e.Exponent)

	if !replaceable(exponent, super) {
		w.Write(e.Raw)
		return rest, nil
	}

	// Write base as-is.
	w.Write(e.Base)

	// Perform replacement and write out.
	replacerunes(exponent, super)
	for _, r := range exponent {
		w.WriteRune(r)
	}

	return rest, nil
}

func parseexp(b []byte) (*Exponentiation, []byte, error) {
	// Find the caret.
	caret := bytes.IndexByte(b, '^')
	if caret < 0 {
		return nil, nil, errors.New("expected caret")
	}

	// Find the end.
	end := bytes.IndexFunc(b, func(r rune) bool {
		return r == '.' || r == ',' || unicode.IsSpace(r)
	})
	if end < 0 {
		return nil, nil, errors.New("expected whitespace")
	}
	if end < caret {
		return nil, nil, errors.New("unexpected whitespace before caret")
	}

	// Construct the parsed exponentiation expression.
	e := &Exponentiation{
		Base:     b[:caret],
		Exponent: unbrace(b[caret+1 : end]),
		Raw:      b[:end],
	}

	return e, b[end:], nil
}

// prefix reports whether b starts with p.
func prefix(b []byte, p string) bool {
	return bytes.HasPrefix(b, []byte(p))
}

// replaceable returns whether every rune in rs has a replacement in repl.
func replaceable(rs []rune, repl map[rune]rune) bool {
	for _, r := range rs {
		if _, ok := repl[r]; !ok {
			return false
		}
	}
	return true
}

// replacerunes replaces runes in rs according to the replacement map.
func replacerunes(rs []rune, repl map[rune]rune) {
	for i := range rs {
		rs[i] = repl[rs[i]]
	}
}

// unbrace removes outer braces, if present.
func unbrace(b []byte) []byte {
	return trimwrap(b, '{', '}')
}

// trimwrap removes open and closing characters, if present.
func trimwrap(b []byte, open, close byte) []byte {
	n := len(b)
	if n >= 2 && b[0] == open && b[n-1] == close {
		b = b[1 : n-1]
	}
	return b
}
