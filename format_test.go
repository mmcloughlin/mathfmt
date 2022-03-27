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
		{Name: "basic_latex_symbol", Input: `x \oplus y`, Expect: "x ⊕ y"},
		{Name: "multi_symbols", Input: "2 <= x <= 10", Expect: "2 ⩽ x ⩽ 10"},

		// Super/subscripts.
		{Name: "sup_brace_replaceable", Input: "x^{i+j}ab", Expect: "xⁱ⁺ʲab"},
		{Name: "sup_numeral_replaceable", Input: "x^123a", Expect: "x¹²³a"},
		{Name: "sup_char_replaceable", Input: "x^ijk", Expect: "x^ijk"},

		{Name: "sup_brace_nonreplaceable", Input: "x^{p+q}pq", Expect: "x^{p+q}pq"},
		{Name: "sup_char_nonreplaceable", Input: "x^qrs", Expect: "x^qrs"},

		{Name: "sub_brace_replaceable", Input: "x_{i+j}ab", Expect: "xᵢ₊ⱼab"},
		{Name: "sub_digit_brace_replaceable", Input: "2_{i+j}ab", Expect: "2_{i+j}ab"},
		{Name: "sub_numeral_boundary_replaceable", Input: "x_123 a", Expect: "x₁₂₃ a"},
		{Name: "sub_numeral_non_boundary", Input: "x_123a", Expect: "x_123a"},
		{Name: "sub_char_replaceable", Input: "x_ijk", Expect: "x_ijk"},

		{Name: "sub_brace_nonreplaceable", Input: "x_{w+x}wx", Expect: "x_{w+x}wx"},
		{Name: "sub_char_nonreplaceable", Input: "x_wxy", Expect: "x_wxy"},

		// Combination of symbols and super/subscripts.
		{Name: "sup_with_symbol", Input: `\oplus^23`, Expect: "⊕²³"},
		{Name: "sub_with_symbol", Input: `\oplus_23`, Expect: "⊕₂₃"},
		{Name: "sup_brace_with_symbol", Input: `\oplus^{i+j}`, Expect: "⊕ⁱ⁺ʲ"},
		{Name: "sub_brace_with_symbol", Input: `\oplus_{i+j}`, Expect: "⊕ᵢ₊ⱼ"},

		// Malformed.
		{Name: "sup_first_char", Input: "^a", Expect: "^a"},
		{Name: "sub_first_char", Input: "_a", Expect: "_a"},

		{Name: "sup_last_char", Input: "a^", Expect: "a^"},
		{Name: "sub_last_char", Input: "a_", Expect: "a_"},

		{Name: "sup_space_before", Input: "pre ^a", Expect: "pre ^a"},
		{Name: "sub_space_before", Input: "pre _a", Expect: "pre _a"},

		// Regression.
		{
			Name:   "sup_with_minus",
			Input:  "2^32-1",
			Expect: "2³²-1",
		},
		{
			Name:   "exp_with_minus",
			Input:  "p256Invert calculates |out| = |in|^{-1}",
			Expect: "p256Invert calculates |out| = |in|⁻¹",
		},
		{
			Name:   "variance",
			Input:  `\sigma^2 = \sum (x_{i} - \mu)^2 / N`,
			Expect: `σ² = ∑ (xᵢ - μ)² / N`,
		},
		{
			Name:   "zeta_function",
			Input:  `\zeta(s) = \sum 1/n^{s}`,
			Expect: `ζ(s) = ∑ 1/nˢ`,
		},
		{
			Name:   "issue_14_eta_sub_2",
			Input:  `\eta_2`,
			Expect: "η₂",
		},
		{
			Name:   "issue_14_eta_sup_2",
			Input:  `\eta^2`,
			Expect: "η²",
		},
	}
	for _, c := range cases {
		c := c // scopelint
		t.Run(c.Name, func(t *testing.T) {
			AssertFormulaOutput(t, c.Input, c.Expect)
		})
	}
}

func TestFormulaNoChange(t *testing.T) {
	// Regression tests for inputs that should have been left alone.
	cases := []string{
		// golang.org/x/crypto
		"\"_acme-challenge\" name of the domain being validated.",                                                            // subscript "_a"
		"echo -n cert | base64 | tr -d '=' | tr '/+' '_-'",                                                                   // subscript "_-"
		"thumbprint is precomputed for testKeyEC in jws_test.go",                                                             // subscript "_t"
		"The \"signature_algorithms\" extension, if present, limits the key exchange",                                        // subscript "_a"
		"testGetCertificate_tokenCache tests the fallback of token certificate fetches",                                      // subscript "_t"
		"https://en.wikipedia.org/wiki/Automated_Certificate_Management_Environment#CAs_&_PKIs_that_offer_ACME_certificates", // subscripts in URL
		"g8TuAS9g5zhq8ELQ3kmjr-KV86GAMgI6VAcGlq3QrzpTCf_30Ab7-zawrfRaFON",                                                    // subscript "_30"
		"JAumQ_I2fjj98_97mk3ihOY4AgVdCDj1z_GCoZkG5Rq7nbCGyosyKWyDX00Zs-n",                                                    // subscript "_97"
		"xiToPMinus1Over3 is ξ^((p-1)/3) where ξ = i+3.",                                                                     // superscript "^("
		"FrobeniusP2 computes (xτ²+yτ+z)^(p²) = xτ^(2p²) + yτ^(p²) + z",                                                      // superscript "^("
		"x for a moment, then after applying the Frobenius, we have x̄ω^(2p)",                                                // superscript "^("
		"x̄ξ^((p-1)/3)ω² and applying the inverse isomorphism eliminates the",                                                // superscript "^("
		"be called when the vector facility is available. Implementation in asm_s390x.s.",                                    // subscript "_s"
		"[1] http://csrc.nist.gov/publications/drafts/fips-202/fips_202_draft.pdf",                                           // subscript "_202"
		"Cert generated by ssh-keygen OpenSSH_6.8p1 OS X 10.10.3",                                                            // subscript "_6"

		// Standard library.
		"     x, ok := <-c",
		"	//	\"->\" == 2",

		// Unhanded cases.
		// "------------------+--------+-----------+----------",
		// "Look for //gdb-<tag>=(v1,v2,v3) and print v1, v2, v3",
		// "====================================================",
		// `  * UNC paths                              (e.g \\server\share\foo\bar)`,
		// `  * absolute paths                         (e.g C:\foo\bar)`,
		// `  * relative paths begin with drive letter (e.g C:foo\bar, C:..\foo\bar, C:.., C:.)`,
		// `  * relative paths begin with '\'          (e.g \foo\bar)`,
		// `  * relative paths begin without '\'       (e.g foo\bar, ..\foo\bar, .., .)`,
	}
	for _, input := range cases {
		AssertFormulaOutput(t, input, input)
	}
}

func AssertFormulaOutput(t *testing.T, input, expect string) {
	t.Helper()
	if got := formula(input); got != expect {
		t.Logf("input  = %q", input)
		t.Logf("got    = %q", got)
		t.Logf("expect = %q", expect)
		t.Fail()
	}
}
