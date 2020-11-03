package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"regexp"
	"strings"
	"unicode"

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

	// Process every comment as a formula.
	transformed := commentreplace(f, formula)

	// Format.
	buf := bytes.NewBuffer(nil)
	if err := format.Node(buf, fset, transformed); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// commentreplace applies repl function to the text of every comment under the root AST.
func commentreplace(root ast.Node, repl func(string) string) ast.Node {
	return astutil.Apply(root, func(c *astutil.Cursor) bool {
		switch n := c.Node().(type) {
		case *ast.Comment:
			c.Replace(&ast.Comment{
				Slash: n.Slash,
				Text:  repl(n.Text),
			})
		case *ast.File:
			for _, g := range n.Comments {
				for _, comment := range g.List {
					comment.Text = repl(comment.Text)
				}
			}
		}
		return true
	}, nil)
}

// Fixed data structures required for formula processing.
var (
	// Symbol replacer.
	replacer *strings.Replacer

	// Regular expressions for super/subscripts.
	supregexp *regexp.Regexp
	subregexp *regexp.Regexp

	// Rune replacement maps.
	super = map[rune]rune{}
	sub   = map[rune]rune{}
)

func init() {
	// Build symbol replacer.
	var oldnew []string
	for symbol, r := range symbols {
		oldnew = append(oldnew, symbol, string([]rune{r}))
	}
	replacer = strings.NewReplacer(oldnew...)

	// Build super/subscript character classes and replacement maps.
	var superclass, subclass []rune
	for _, char := range chars {
		if char.Super != None {
			superclass = append(superclass, char.Char)
			super[char.Char] = char.Super
		}
		if char.Sub != None {
			subclass = append(subclass, char.Char)
			sub[char.Char] = char.Sub
		}
	}

	// Build regular expressions.
	supregexp = regexp.MustCompile(`(\b[A-Za-z0-9]|[)\pL\pS^A-Za-z0-9])\^(\d+|\{` + charclass(superclass) + `+\}|` + charclass(superclass) + `\s)`)
	subregexp = regexp.MustCompile(`(\b[A-Za-z]|\pS|\p{Greek})_(\d+\b|\{` + charclass(subclass) + `+\})`)
}

// charclass builds a regular expression character class from a list of runes.
func charclass(runes []rune) string {
	return strings.ReplaceAll("["+string(runes)+"]", "-", `\-`)
}

// formula processes a formula in s, writing the result to w.
func formula(s string) string {
	// Replace symbols.
	s = replacer.Replace(s)

	// Replace superscripts.
	s = supregexp.ReplaceAllStringFunc(s, subsupreplacer(super))

	// Replace subscripts.
	s = subregexp.ReplaceAllStringFunc(s, subsupreplacer(sub))

	return s
}

// subsupreplacer builds a replacement function that applies the repl rune map
// to a matched super/subscript.
func subsupreplacer(repl map[rune]rune) func(string) string {
	return func(s string) string {
		var runes []rune
		for i, r := range s {
			if i == 0 || unicode.IsSpace(r) {
				runes = append(runes, r)
			} else if repl[r] != None {
				runes = append(runes, repl[r])
			}
		}
		return string(runes)
	}
}
