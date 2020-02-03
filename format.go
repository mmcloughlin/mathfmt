package main

import (
	"bytes"
	"errors"
	"strings"
	"unicode"
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

// macro processes a macro starting at r. Note r points at the character directly after the macro name.
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

// formula processes a formula in r, writing the result to w.
func formula(w *bytes.Buffer, s string) error {
	if len(s) == 0 {
		return nil
	}

	// Replace symbols.
	s = replacer.Replace(s)

	// Replace super/subscripts.
	r := []rune(s)
	last := None
	for len(r) > 0 {
		// Look for a super/subscript character.
		var repl map[rune]rune
		switch r[0] {
		case '^':
			repl = super
		case '_':
			repl = sub
		default:
			w.WriteRune(r[0])
			last = r[0]
			r = r[1:]
			continue
		}

		// Perform replacement.
		if unicode.IsPrint(last) && !unicode.IsSpace(last) {
			var err error
			r, err = supsub(w, r, repl)
			if err != nil {
				return err
			}
		} else {
			w.WriteRune(r[0])
			r = r[1:]
		}

		last = None
	}

	return nil
}

func supsub(w *bytes.Buffer, r []rune, repl map[rune]rune) ([]rune, error) {
	arg, rest, err := parsearg(r[1:])
	if err != nil {
		return nil, err
	}

	// If we could not parse an argument, or its not replaceable, just write the
	// sub/script operator and return.
	if len(arg) == 0 || !replaceable(arg, repl) {
		w.WriteRune(r[0])
		return r[1:], nil
	}

	// Perform the replacement.
	replacerunes(arg, repl)
	w.WriteString(string(arg))

	return rest, nil
}

func parsearg(r []rune) ([]rune, []rune, error) {
	if len(r) == 0 {
		return nil, r, nil
	}

	// Braced.
	if r[0] == '{' {
		arg, rest, err := parsebraces(string(r))
		if err != nil {
			return nil, nil, err
		}
		return []rune(arg[1 : len(arg)-1]), []rune(rest), nil
	}

	// Numeral.
	i := 0
	for ; i < len(r) && unicode.IsNumber(r[i]); i++ {
	}
	if i > 0 {
		return r[:i], r[i:], nil
	}

	// Default to just one character.
	return r[:1], r[1:], nil
}

// parsebraces parses matching braces starting at the beginning of r.
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

// replaceable returns whether every rune in rs has a replacement in repl.
func replaceable(r []rune, repl map[rune]rune) bool {
	for _, c := range r {
		if _, ok := repl[c]; !ok {
			return false
		}
	}
	return true
}

// replacerunes replaces runes in rs according to the replacement map.
func replacerunes(r []rune, repl map[rune]rune) {
	for i := range r {
		r[i] = repl[r[i]]
	}
}
