// +build ignore

package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

func main() {
	os.Exit(main1())
}

func main1() int {
	if err := mainerr(); err != nil {
		log.Print(err)
		return 1
	}
	return 0
}

var (
	input      = flag.String("input", "unimathsymbols.txt", "unimathsymbols database")
	outputtype = flag.String("type", "table", "output type")
	output     = flag.String("output", "", "output file (default stdout)")
)

var generators = map[string]Generator{
	"table": GoTable("main", "symbols"),
	"doc":   Documentation,
}

var aliasmap = map[string]string{
	"+-":  `\pm`,
	"-+":  `\mp`,
	"==":  `\equiv`,
	"<=":  `\leqslant`,
	">=":  `\geqslant`,
	"||":  `\parallel`,
	"<-":  `\leftarrow`,
	"->":  `\rightarrow`,
	"|->": `\mapsto`,
}

func mainerr() error {
	// Parse flags.
	flag.Parse()

	g := generators[*outputtype]
	if g == nil {
		return fmt.Errorf("unknown output type %q", *outputtype)
	}

	// Load symbols.
	symbols, err := LoadSymbolsFile(*input)
	if err != nil {
		return err
	}

	// Process into macros.
	macros := MacrosFromSymbols(symbols)

	aliases, err := BuildAliases(macros, aliasmap)
	if err != nil {
		return err
	}

	macros = append(macros, aliases...)
	sort.SliceStable(macros, func(i, j int) bool {
		return macros[i].Char < macros[j].Char
	})

	// Generate output.
	b, err := g(macros)
	if err != nil {
		return err
	}

	// Write.
	if *output != "" {
		return ioutil.WriteFile(*output, b, 0666)
	}
	_, err = os.Stdout.Write(b)
	return err
}

// Symbol from the unimathsymbols database.
type Symbol struct {
	Char               rune
	LaTeXCommand       string
	UnicodeMathCommand string
	UnicodeMathClass   string
	TeXCategory        string
	Requirements       []string
	Conflicts          []string
	Aliases            []string
	Approx             []string
	SeeAlso            []string
	TextMode           []string
	Comments           []string
	CharacterName      string
}

// LoadSymbolsFile reads symbols from the given filename.
func LoadSymbolsFile(filename string) (symbols []Symbol, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		if errc := f.Close(); errc != nil && err == nil {
			err = errc
		}
	}()
	return LoadSymbols(f)
}

// LoadSymbols from the reader r.
func LoadSymbols(r io.Reader) ([]Symbol, error) {
	var symbols []Symbol
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Text()

		// Skip comments.
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		// Parse symbol.
		symbol, err := parsesymbol(line)
		if err != nil {
			return nil, err
		}

		symbols = append(symbols, symbol)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return symbols, nil
}

// parsesymbol parses a single line of the unimathsymbols database.
func parsesymbol(line string) (Symbol, error) {
	// Break into fields.
	fields := strings.Split(line, "^")
	if len(fields) != 8 {
		return Symbol{}, errors.New("symbol line must have 8 fields")
	}

	symbol := Symbol{}

	// Code point (Unicode character number)
	//
	//    There may be more than one record for one code point,
	//    if there are different TeX commands for the same character.
	//    (changed 2015-09-21, before the code point was unique.)

	codepoint, err := strconv.ParseUint(fields[0], 16, 32)
	if err != nil {
		return Symbol{}, err
	}
	symbol.Char = rune(codepoint)

	// Literal character (UTF-8 encoded)

	// We already have this from the codepoint.

	// (La)TeX _`command`
	//
	//    Preferred representation of the character in (La)TeX.
	//    Alternative commands are listed in the comments_ field.

	symbol.LaTeXCommand = fields[2]
	symbol.UnicodeMathCommand = fields[3]

	// Unicode math character class (after MathClassEx_).
	//
	//    .. _MathClassEx:
	//       http://www.unicode.org/Public/math/revision-11/MathClassEx-11.txt
	//
	//    The class can be one of:
	//
	//    :N: Normal- includes all digits and symbols requiring only one form
	//    :A: Alphabetic
	//    :B: Binary
	//    :C: Closing – usually paired with opening delimiter
	//    :D: Diacritic
	//    :F: Fence - unpaired delimiter (often used as opening or closing)
	//    :G: Glyph_Part- piece of large operator
	//    :L: Large -n-ary or Large operator, often takes limits
	//    :O: Opening – usually paired with closing delimiter
	//    :P: Punctuation
	//    :R: Relation- includes arrows
	//    :S: Space
	//    :U: Unary – operators that are only unary
	//    :V: Vary – operators that can be unary or binary depending on context
	//    :X: Special –characters not covered by other classes
	//
	//    C, O, and F operators are stretchy. In addition some binary
	//    operators, such as 002F are stretchy as noted in the descriptive
	//    comments. The classes are also useful in determining extra spacing
	//    around the operators as discussed in UTR#25.

	symbol.UnicodeMathClass = fields[4]

	// TeX math category (after unimath-symbols_)
	//
	//    .. _unimath-symbols:
	//       http://mirror.ctan.org/macros/latex/contrib/unicode-math/unimath-symbols.pdf

	symbol.TeXCategory = fields[5]

	// Requirements and Conflicts
	//
	//    Space delimited list of LaTeX packages or features [1]_ providing
	//    the LaTeX command_ or conflicting with it.
	//
	//    Packages/features preceded by a HYPHEN-MINUS (-) use the command
	//    for a different character or purpose.
	//
	//    To save space, packages providing/modifying (almost) all commands
	//    of a feature or another package are not listed here but in the
	//    ``packages.txt`` file.
	//
	//    .. [1] A feature can be a set of commands common to several packages,
	//    	    (e.g. ``mathbb`` or ``slantedGreek``) or a constraint (e.g.
	//	    ``literal`` mapping plain characters to upright face).

	for _, feature := range strings.Fields(fields[6]) {
		if feature[0] == '-' {
			symbol.Conflicts = append(symbol.Conflicts, feature[1:])
		} else {
			symbol.Requirements = append(symbol.Requirements, feature)
		}
	}

	// Descriptive _`comments`
	//
	//    The descriptive comments provide more information about the
	//    character, or its specific appearance or use.
	//
	//    Some descriptions contain references to related commands,
	//    marked by a character describing the relation
	//
	//    :=:  equals  (alias commands),
	//    :#:  approx  (compat mapping, different character with same glyph),
	//    :x:  → cross reference/see also (related, false friends, and name clashes),
	//    :t:  text    (text mode command),
	//
	//    followed by requirements in parantheses, and
	//    delimited by commas.
	//
	//    Comments in UPPERCASE are Unicode character names

	for _, part := range strings.Split(fields[7], ",") {
		part = strings.TrimSpace(part)
		switch {
		case strings.HasPrefix(part, "= "):
			symbol.Aliases = append(symbol.Aliases, part[2:])
		case strings.HasPrefix(part, "# "):
			symbol.Approx = append(symbol.Approx, part[2:])
		case strings.HasPrefix(part, "x "):
			symbol.SeeAlso = append(symbol.SeeAlso, part[2:])
		case strings.HasPrefix(part, "t "):
			symbol.TextMode = append(symbol.TextMode, part[2:])
		case strings.ToUpper(part) == part:
			symbol.CharacterName = part
		default:
			symbol.Comments = append(symbol.Comments, part)
		}
	}

	return symbol, nil
}

// Macro is a symbol macro that will be applied by mathfmt.
type Macro struct {
	Command       string
	Char          rune
	Section       string
	CharacterName string
}

// MacrosFromSymbols filters the list of symbols down to a list of macros. When
// multiple symbols have the same preferred command, the first in order will be
// kept.
func MacrosFromSymbols(symbols []Symbol) []Macro {
	macros := make([]Macro, 0, len(symbols))
	seen := map[string]bool{}
	for _, symbol := range symbols {
		if !IncludeSymbol(symbol) {
			continue
		}
		cmd := SymbolCommand(symbol)
		if !seen[cmd] {
			macros = append(macros, Macro{
				Command:       cmd,
				Char:          symbol.Char,
				Section:       symbol.TeXCategory,
				CharacterName: symbol.CharacterName,
			})
			seen[cmd] = true
		}
	}
	return macros
}

// IncludeSymbol reports whether symbol should be in mathfmt.
func IncludeSymbol(symbol Symbol) bool {
	// Skip unprintable characters (various types of spaces and invisible characters).
	if !unicode.IsPrint(symbol.Char) {
		return false
	}

	// Single character commands don't serve any purpose.
	cmd := SymbolCommand(symbol)
	return len(cmd) >= 2
}

// SymbolCommand returns the preferred command for a symbol.
func SymbolCommand(s Symbol) string {
	if s.LaTeXCommand != "" {
		return s.LaTeXCommand
	}
	return s.UnicodeMathCommand
}

// aliassection is the section to use for alias macros.
const aliassection = "alias"

// BuildAliases builds a list of alias macros for the given from -> to mapping.
func BuildAliases(macros []Macro, aliasmap map[string]string) ([]Macro, error) {
	// Map from cmd to macro.
	bycmd := map[string]Macro{}
	for _, m := range macros {
		bycmd[m.Command] = m
	}

	aliases := make([]Macro, 0, len(aliasmap))
	for from, to := range aliasmap {
		alias, ok := bycmd[to]
		if !ok {
			return nil, fmt.Errorf("unknown macro %q", to)
		}

		alias.Command = from
		alias.Section = aliassection
		aliases = append(aliases, alias)
	}

	return aliases, nil
}

// Generator generates output from a macros list.
type Generator func([]Macro) ([]byte, error)

// GoTable generates a go source file with a mapping from command to rune.
func GoTable(pkg, varname string) Generator {
	return Generator(func(macros []Macro) ([]byte, error) {
		buf := bytes.NewBuffer(nil)
		_, self, _, _ := runtime.Caller(0)
		fmt.Fprintf(buf, "// Code generated by %s. DO NOT EDIT.\n\n", filepath.Base(self))
		fmt.Fprintf(buf, "package %s\n\n", pkg)
		fmt.Fprintf(buf, "var %s = map[string]rune{\n", varname)
		for _, m := range macros {
			fmt.Fprintf(buf, "\t%#+q: %+q,\n", m.Command, m.Char)
		}
		fmt.Fprint(buf, "}\n")
		return format.Source(buf.Bytes())
	})
}

// sections in the documentation.
var sections = []struct {
	ID   string
	Name string
}{
	{aliassection, "Aliases"},
	{"mathopen", "Opening Symbols"},  // 1
	{"mathclose", "Closing Symbols"}, // 2
	{"mathfence", "Fence Symbols"},   // 3
	{"mathover", "Over Symbols"},     // 5
	{"mathunder", "Under Symbols"},   // 6
	{"mathaccent", "Accents"},        // 7
	{"mathop", "Big Operators"},      // 9
	{"mathradical", "Radicals"},
	{"mathbin", "Binary relations"},       // 10
	{"mathord", "Ordinary Symbols"},       // 11
	{"mathrel", "Relation Symbols"},       // 12
	{"mathalpha", "Alphabetical Symbols"}, // 13
}

// Documentation generates markdown documentation for the symbol replacements.
func Documentation(macros []Macro) ([]byte, error) {
	// Set of defined sections.
	defined := map[string]bool{}
	for _, section := range sections {
		defined[section.ID] = true
	}

	// Divide by section.
	sectionmacros := map[string][]Macro{}
	for _, m := range macros {
		if !defined[m.Section] {
			return nil, fmt.Errorf("unknown section %q", m.Section)
		}
		sectionmacros[m.Section] = append(sectionmacros[m.Section], m)
	}

	// Output.
	buf := bytes.NewBuffer(nil)
	fmt.Fprint(buf, "# Symbols Reference\n")
	for _, section := range sections {
		if len(sectionmacros[section.ID]) == 0 {
			return nil, fmt.Errorf("no macros in section %q", section.ID)
		}
		fmt.Fprintf(buf, "\n## %s\n\n", section.Name)

		fmt.Fprint(buf, "| Char | Command | Character Name |\n")
		fmt.Fprint(buf, "| --- | --- | --- |\n")
		for _, m := range sectionmacros[section.ID] {
			fmt.Fprintf(buf, "| `%c` ", m.Char)
			fmt.Fprintf(buf, "| `%s` ", strings.ReplaceAll(m.Command, "|", `\|`))
			fmt.Fprintf(buf, "| %s ", m.CharacterName)
			fmt.Fprint(buf, "|\n")
		}
	}
	return buf.Bytes(), nil
}
