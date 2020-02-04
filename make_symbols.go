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

	// Postprocess.
	symbols = FilterSymbols(symbols)

	// Generate output.
	b, err := g(symbols)
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

func LoadSymbolsFile(filename string) ([]Symbol, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadSymbols(f)
}

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

// FilterSymbols filters the list of symbols to those we want to include in mathfmt.
func FilterSymbols(symbols []Symbol) []Symbol {
	selected := make([]Symbol, 0, len(symbols))
	for _, symbol := range symbols {
		if IncludeSymbol(symbol) {
			selected = append(selected, symbol)
		}
	}
	return selected
}

// Include reports whether symbol should be in mathfmt.
func IncludeSymbol(symbol Symbol) bool {
	str := string([]rune{symbol.Char})

	// Skip unprintable characters (various types of spaces and invisible characters).
	if !unicode.IsPrint(symbol.Char) {
		return false
	}

	// We want symbols with a command, that's not just the same as the character itself.
	cmd := SymbolCommand(symbol)
	if cmd == "" || str == cmd {
		return false
	}

	return true
}

// SymbolCommand returns the preferred command for a symbol.
func SymbolCommand(s Symbol) string {
	if s.LaTeXCommand != "" {
		return s.LaTeXCommand
	}
	return s.UnicodeMathCommand
}

// Generator generates output from a symbol list.
type Generator func([]Symbol) ([]byte, error)

// GoTable generates a go source file with a mapping from command to rune.
func GoTable(pkg, varname string) Generator {
	return Generator(func(symbols []Symbol) ([]byte, error) {
		buf := bytes.NewBuffer(nil)
		_, self, _, _ := runtime.Caller(0)
		fmt.Fprintf(buf, "// Code generated by %s. DO NOT EDIT.\n\n", filepath.Base(self))
		fmt.Fprintf(buf, "package %s\n\n", pkg)
		fmt.Fprintf(buf, "var %s = map[string]rune{\n", varname)
		for _, symbol := range symbols {
			fmt.Fprintf(buf, "\t%#+q: %+q,\n", SymbolCommand(symbol), symbol.Char)
		}
		fmt.Fprint(buf, "}\n")
		return format.Source(buf.Bytes())
	})
}
