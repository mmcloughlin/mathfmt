package main

import (
	"testing"
)

func TestFormula(t *testing.T) {
	cases := []struct {
		Name   string
		Input  string
		Expect string
	}{
		{Name: "empty", Input: "", Expect: ""},
		{Name: "no_math_chars", Input: "Hello, World!", Expect: "Hello, World!"},
		{Name: "nested_braces", Input: "{{}{{}{}}}", Expect: "{{}{{}{}}}"},

		// Symbols.
		{Name: "basic_symbol", Input: "x +- y", Expect: "x ± y"},
		{Name: "basic_latex_symbol", Input: "x \\oplus y", Expect: "x ⊕ y"},
		{Name: "multi_symbols", Input: "2 <= x <= 10", Expect: "2 ⩽ x ⩽ 10"},

		// Super/subscripts.
		{Name: "sup_brace_replaceable", Input: "x^{i+j}ab", Expect: "xⁱ⁺ʲab"},
		{Name: "sup_numeral_replaceable", Input: "x^123a", Expect: "x¹²³a"},
		{Name: "sup_char_replaceable", Input: "x^ijk", Expect: "xⁱjk"},

		{Name: "sup_brace_nonreplaceable", Input: "x^{p+q}pq", Expect: "x^{p+q}pq"},
		{Name: "sup_char_nonreplaceable", Input: "x^qrs", Expect: "x^qrs"},

		{Name: "sub_brace_replaceable", Input: "x_{i+j}ab", Expect: "xᵢ₊ⱼab"},
		{Name: "sub_numeral_replaceable", Input: "x_123a", Expect: "x₁₂₃a"},
		{Name: "sub_char_replaceable", Input: "x_ijk", Expect: "xᵢjk"},

		{Name: "sub_brace_nonreplaceable", Input: "x_{w+x}wx", Expect: "x_{w+x}wx"},
		{Name: "sub_char_nonreplaceable", Input: "x_wxy", Expect: "x_wxy"},

		// Combination.
		{Name: "sup_with_symbol", Input: "\\oplus^23", Expect: "⊕²³"},
		{Name: "sub_with_symbol", Input: "\\oplus_23", Expect: "⊕₂₃"},

		// Malformed.
		{Name: "sup_first_char", Input: "^a", Expect: "^a"},
		{Name: "sub_first_char", Input: "_a", Expect: "_a"},

		{Name: "sup_last_char", Input: "a^", Expect: "a^"},
		{Name: "sub_last_char", Input: "a_", Expect: "a_"},

		{Name: "sup_space_before", Input: "pre ^a", Expect: "pre ^a"},
		{Name: "sub_space_before", Input: "pre _a", Expect: "pre _a"},

		{Name: "sup_consecutive", Input: "pre ^^^^^^^a post", Expect: "pre ^^^^^^^a post"},
		{Name: "sub_consecutive", Input: "pre _______a post", Expect: "pre _______a post"},

		// Regression.
		{Name: "sup_with_minus", Input: "2^32-1", Expect: "2³²-1"},
	}
	for _, c := range cases {
		c := c // scopelint
		t.Run(c.Name, func(t *testing.T) {
			got, err := formula(c.Input)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.Expect {
				t.Logf("input  = %q", c.Input)
				t.Logf("got    = %q", got)
				t.Logf("expect = %q", c.Expect)
				t.FailNow()
			}
		})
	}
}
