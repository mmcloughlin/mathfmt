package main

import (
	"bytes"
	"errors"
	"unicode"
)

// Source processes the source code in b.
func Source(b []byte) ([]byte, error) {
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
		switch {
		// Stop at new line.
		case b[0] == '\n':
			w.WriteByte(b[0])
			b = b[1:]
			return b, nil
		// Is this the start of an exponent?
		case prefix(b, "2^"):
			rest, err := exp(w, b)
			if err != nil {
				return nil, err
			}
			b = rest
		default:
			w.WriteByte(b[0])
			b = b[1:]
		}
	}
	return b, nil
}

// exp processes an exponentiation.
func exp(w *bytes.Buffer, b []byte) ([]byte, error) {
	// Find the caret.
	caret := bytes.IndexByte(b, '^')
	if caret < 0 {
		return nil, errors.New("expected caret")
	}

	// Find the end.
	end := bytes.IndexFunc(b, unicode.IsSpace)
	if end < 0 {
		return nil, errors.New("expected whitespace")
	}
	if end < caret {
		return nil, errors.New("unexpected whitespace before caret")
	}

	// Is the exponent replaceable with superscripts? If not write out unchanged and return.
	e := bytes.Runes(b[caret+1 : end])

	if !replaceable(e, super) {
		w.Write(b[:end])
		return b[end:], nil
	}

	// Write up to the caret as-is.
	w.Write(b[:caret])

	// Perform replacement and write out.
	replacerunes(e, super)
	for _, r := range e {
		w.WriteRune(r)
	}

	return b[end:], nil
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
