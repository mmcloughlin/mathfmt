package main

// None is the zero rune.
var None rune

// Character represents a character with its superscript and subscript variants.
type Character struct {
	Char  rune
	Super rune
	Sub   rune
}

// chars is the table of super/subscriptable characters.
var chars = []Character{
	{'0', '\u2070', '\u2080'},
	{'1', '\u00B9', '\u2081'},
	{'2', '\u00B2', '\u2082'},
	{'3', '\u00B3', '\u2083'},
	{'4', '\u2074', '\u2084'},
	{'5', '\u2075', '\u2085'},
	{'6', '\u2076', '\u2086'},
	{'7', '\u2077', '\u2087'},
	{'8', '\u2078', '\u2088'},
	{'9', '\u2079', '\u2089'},
	{'a', '\u1d43', '\u2090'},
	{'b', '\u1d47', None},
	{'c', '\u1d9c', None},
	{'d', '\u1d48', None},
	{'e', '\u1d49', '\u2091'},
	{'f', '\u1da0', None},
	{'g', '\u1d4d', None},
	{'h', '\u02b0', '\u2095'},
	{'i', '\u2071', '\u1d62'},
	{'j', '\u02b2', '\u2c7c'},
	{'k', '\u1d4f', '\u2096'},
	{'l', '\u02e1', '\u2097'},
	{'m', '\u1d50', '\u2098'},
	{'n', '\u207f', '\u2099'},
	{'o', '\u1d52', '\u2092'},
	{'p', '\u1d56', '\u209a'},
	{'q', None, None},
	{'r', '\u02b3', '\u1d63'},
	{'s', '\u02e2', '\u209b'},
	{'t', '\u1d57', '\u209c'},
	{'u', '\u1d58', '\u1d64'},
	{'v', '\u1d5b', '\u1d65'},
	{'w', '\u02b7', None},
	{'x', '\u02e3', '\u2093'},
	{'y', '\u02b8', None},
	{'z', None, None},
	{'A', '\u1d2c', None},
	{'B', '\u1d2e', None},
	{'C', None, None},
	{'D', '\u1d30', None},
	{'E', '\u1d31', None},
	{'F', None, None},
	{'G', '\u1d33', None},
	{'H', '\u1d34', None},
	{'I', '\u1d35', None},
	{'J', '\u1d36', None},
	{'K', '\u1d37', None},
	{'L', '\u1d38', None},
	{'M', '\u1d39', None},
	{'N', '\u1d3a', None},
	{'O', '\u1d3c', None},
	{'P', '\u1d3e', None},
	{'Q', None, None},
	{'R', '\u1d3f', None},
	{'S', None, None},
	{'T', '\u1d40', None},
	{'U', '\u1d41', None},
	{'V', '\u2c7d', None},
	{'W', '\u1d42', None},
	{'X', None, None},
	{'Y', None, None},
	{'Z', None, None},
	{'+', '\u207A', '\u208A'},
	{'-', '\u207B', '\u208B'},
	{'=', '\u207C', '\u208C'},
	{'(', '\u207D', '\u208D'},
	{')', '\u207E', '\u208E'},
}

// symbols defines symbol replacements.
var symbols = map[string]rune{
	"+-":      '\u00B1', // PLUS-MINUS SIGN
	"-+":      '\u2213', // MINUS-OR-PLUS SIGN
	"==":      '\u2261', // IDENTICAL TO
	"<=":      '\u2A7D', // LESS-THAN OR SLANTED EQUAL TO
	">=":      '\u2A7E', // GREATER-THAN OR SLANTED EQUAL TO
	"||":      '\u2225', // PARALLEL TO
	"<-":      '\u2191', // LEFTWARDS ARROW
	"->":      '\u2192', // RIGHTWARDS ARROW
	"|->":     '\u21a6', // RIGHTWARDS ARROW FROM BAR
	"\\oplus": '\u2295', // CIRCLED PLUS
}
