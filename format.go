package main

import (
	"bytes"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

// macroname is the name of the macro that applies math formatting.
const macroname = "\\mathfmt"

// Format processes the source code in b.
func Format(b []byte) ([]byte, error) {
	var buf bytes.Buffer
	s := string(b)
	for len(s) > 0 {
		// Look for the next macro.
		i := strings.Index(s, macroname)

		// Exit if not found.
		if i < 0 {
			buf.WriteString(s)
			break
		}

		// Write out up to the macro.
		buf.WriteString(s[:i])
		s = s[i:]

		// Process the macro.
		rest, err := macro(&buf, s[len(macroname):])
		if err != nil {
			return nil, err
		}
		s = rest
	}

	return buf.Bytes(), nil
}

// macro processes a macro starting at s. Note s begins at the character directly after the macro name.
func macro(w *bytes.Buffer, s string) (string, error) {
	if len(s) == 0 {
		return "", errors.New("empty macro")
	}

	arg, rest, err := parsebraces(s)
	if err != nil {
		return "", err
	}

	n := len(arg)
	if err := formula(w, arg[1:n-1]); err != nil {
		return "", err
	}

	return rest, nil
}

// Fixed data structures required for formula processing.
var (
	replacer *strings.Replacer // replacer for symbols.
	super    = map[rune]rune{} // replacement map for superscript characters.
	sub      = map[rune]rune{} // replacement map for subscript characters.
)

func init() {
	// Build symbol replacer.
	var oldnew []string
	for symbol, r := range symbols {
		oldnew = append(oldnew, symbol, string([]rune{r}))
	}
	replacer = strings.NewReplacer(oldnew...)

	// Build super/subscript replacement maps.
	for _, char := range chars {
		if char.Super != None {
			super[char.Char] = char.Super
		}
		if char.Sub != None {
			sub[char.Char] = char.Sub
		}
	}
}

// formula processes a formula in s, writing the result to w.
func formula(w *bytes.Buffer, s string) error {
	if len(s) == 0 {
		return nil
	}

	// Replace symbols.
	s = replacer.Replace(s)

	// Replace super/subscripts.
	last := None
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)

		// Look for a super/subscript character.
		var repl map[rune]rune
		switch r {
		case '^':
			repl = super
		case '_':
			repl = sub
		default:
			w.WriteRune(r)
			last = r
			s = s[size:]
			continue
		}

		// Perform replacement.
		if unicode.IsPrint(last) && !unicode.IsSpace(last) {
			var err error
			s, err = supsub(w, s, repl)
			if err != nil {
				return err
			}
		} else {
			w.WriteRune(r)
			s = s[size:]
		}

		last = None
	}

	return nil
}

// supsub processes a super/subscript starting at s, writing the result to w.
// The repl map provides the mapping from runes to the corresponding
// super/subscripted versions. Note the first character of s should be the "^"
// or "_" operator.
func supsub(w *bytes.Buffer, s string, repl map[rune]rune) (string, error) {
	arg, rest, err := parsearg(s[1:])
	if err != nil {
		return "", err
	}

	// If we could not parse an argument, or its not replaceable, just write the
	// sub/script operator and return.
	if len(arg) == 0 || !replaceable(arg, repl) {
		w.WriteByte(s[0])
		return s[1:], nil
	}

	// Perform the replacement.
	for _, r := range arg {
		w.WriteRune(repl[r])
	}

	return rest, nil
}

// parsearg parses the argument to a super/subscript.
func parsearg(s string) (string, string, error) {
	if len(s) == 0 {
		return "", "", nil
	}

	// Braced.
	if s[0] == '{' {
		arg, rest, err := parsebraces(s)
		if err != nil {
			return "", "", err
		}
		return arg[1 : len(arg)-1], rest, nil
	}

	// Look for a numeral.
	i := 0
	for ; i < len(s) && '0' <= s[i] && s[i] <= '9'; i++ {
	}
	if i > 0 {
		return s[:i], s[i:], nil
	}

	// Default to the first rune.
	_, i = utf8.DecodeRuneInString(s)
	return s[:i], s[i:], nil
}

// parsebraces parses matching braces starting at the beginning of s.
func parsebraces(s string) (string, string, error) {
	if len(s) == 0 || s[0] != '{' {
		return "", "", errors.New("expected {")
	}

	depth := 0
	for i, r := range s {
		// Adjust depth if we see open or close brace.
		switch r {
		case '{':
			depth++
		case '}':
			depth--
		}

		// Continue if we have not reached matched braces.
		if depth > 0 {
			continue
		}

		// Return the matched braces.
		return s[:i+1], s[i+1:], nil
	}

	return "", "", errors.New("unmatched braces")
}

// replaceable returns whether every rune in s has a replacement in repl.
func replaceable(s string, repl map[rune]rune) bool {
	for _, r := range s {
		if _, ok := repl[r]; !ok {
			return false
		}
	}
	return true
}
