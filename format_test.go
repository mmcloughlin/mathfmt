package main

import (
	"testing"
)

//func TestParseExp(t *testing.T) {
//	cases := []struct {
//		Input    string
//		Base     string
//		Exponent string
//		Rest     string
//	}{
//		{"2^4 rest", "2", "4", " rest"},
//		{"3^n rest", "3", "n", " rest"},
//		{"50^abc rest", "50", "abc", " rest"},
//		{"2^4\n", "2", "4", "\n"},
//		{"2^4,", "2", "4", ","},
//		{"2^4.rest", "2", "4", ".rest"},
//		{"2^{r-l} rest", "2", "r-l", " rest"},
//		{"x^{a+b}. rest", "x", "a+b", ". rest"},
//	}
//	for _, c := range cases {
//		t.Logf("input: %q", c.Input)
//		e, rest, err := parseexp([]byte(c.Input))
//		if err != nil {
//			t.Error(err)
//			continue
//		}
//		AssertStringEqual(t, "base", string(e.Base), c.Base)
//		AssertStringEqual(t, "exp", string(e.Exponent), c.Exponent)
//		AssertStringEqual(t, "rest", string(rest), c.Rest)
//		AssertStringEqual(t, "raw+rest", string(e.Raw)+string(rest), c.Input)
//	}
//}

func AssertStringEqual(t *testing.T, name, got, expect string) {
	t.Helper()
	if got != expect {
		t.Errorf("%s: got %q expect %q", name, got, expect)
	}
}
