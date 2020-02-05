package main

import (
	"bytes"
	"errors"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/ast/astutil"
)

// Format processes the source code.
func Format(src []byte) ([]byte, error) {
	// Parse.
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Apply transform.
	transformed := CommentTransform(f, func(text string) string {
		newtext, errf := formula(text)
		if errf != nil {
			err = errf
			return text
		}
		return newtext
	})
	if err != nil {
		return nil, err
	}

	// Format.
	buf := bytes.NewBuffer(nil)
	if err := format.Node(buf, fset, transformed); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// CommentTransform applies transform to the text of every comment under the root AST.
func CommentTransform(root ast.Node, transform func(string) string) ast.Node {
	return astutil.Apply(root, func(c *astutil.Cursor) bool {
		switch n := c.Node().(type) {
		case *ast.Comment:
			c.Replace(&ast.Comment{
				Slash: n.Slash,
				Text:  transform(n.Text),
			})
		case *ast.File:
			for _, g := range n.Comments {
				for _, comment := range g.List {
					comment.Text = transform(comment.Text)
				}
			}
		}
		return true
	}, nil)
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
func formula(s string) (string, error) {
	if len(s) == 0 {
		return "", nil
	}

	// Replace symbols.
	s = replacer.Replace(s)

	// Replace super/subscripts.
	buf := bytes.NewBuffer(nil)
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
			buf.WriteRune(r)
			last = r
			s = s[size:]
			continue
		}

		// Perform replacement.
		if unicode.IsPrint(last) && !unicode.IsSpace(last) {
			var err error
			s, err = supsub(buf, s, repl)
			if err != nil {
				return "", err
			}
		} else {
			buf.WriteRune(r)
			s = s[size:]
		}

		last = None
	}

	return buf.String(), nil
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
