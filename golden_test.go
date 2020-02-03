package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
	"unicode"
)

var update = flag.Bool("update", false, "update golden files")

type TestCase struct {
	Name   string
	Input  string
	Golden string
}

func LoadTestCases(tb testing.TB) []TestCase {
	tb.Helper()

	ext := ".in"
	inputs, err := filepath.Glob("testdata/*" + ext)
	if err != nil {
		tb.Fatal(err)
	}

	var cases []TestCase
	for _, input := range inputs {
		noext := strings.TrimSuffix(input, ext)
		cases = append(cases, TestCase{
			Input:  input,
			Name:   filepath.Base(noext),
			Golden: noext + ".golden",
		})
	}

	return cases
}

func TestFormatGolden(t *testing.T) {
	for _, c := range LoadTestCases(t) {
		c := c // scopelint
		t.Run(c.Name, func(t *testing.T) {
			// Read input.
			b, err := ioutil.ReadFile(c.Input)
			if err != nil {
				t.Fatal(err)
			}

			// Format.
			got, err := Format(b)
			if err != nil {
				t.Fatal(err)
			}

			// Update golden file if requested.
			if *update {
				if err := ioutil.WriteFile(c.Golden, got, 0666); err != nil {
					t.Fatal(err)
				}
			}

			// Read golden file.
			expect, err := ioutil.ReadFile(c.Golden)
			if err != nil {
				t.Fatal(err)
			}

			// Compare.
			AssertOutputEquals(t, expect, got)
		})
	}
}

func TestInputsASCII(t *testing.T) {
	for _, c := range LoadTestCases(t) {
		c := c // scopelint
		t.Run(c.Name, func(t *testing.T) {
			// Read input.
			b, err := ioutil.ReadFile(c.Input)
			if err != nil {
				t.Fatal(err)
			}

			// Check for non-ASCII.
			line := 1
			for _, r := range string(b) {
				switch {
				case r == '\n':
					line++
				case r > unicode.MaxASCII:
					t.Errorf("%d: non-ascii character %c", line, r)
				}
			}
		})
	}
}

func AssertOutputEquals(t *testing.T, expect, got []byte) {
	t.Helper()

	if bytes.Equal(expect, got) {
		return
	}
	t.Fail()

	// Break into lines.
	expectlines := strings.Split(string(expect), "\n")
	gotlines := strings.Split(string(got), "\n")

	if len(expectlines) != len(gotlines) {
		t.Fatalf("line number mismatch: got %v expect %v", len(gotlines), len(expectlines))
	}

	for i := range expectlines {
		if expectlines[i] != gotlines[i] {
			t.Errorf("line %d:\n\tgot    = %q\n\texpect = %q", i+1, gotlines[i], expectlines[i])
		}
	}
}

func BenchmarkFormatGolden(b *testing.B) {
	for _, c := range LoadTestCases(b) {
		c := c // scopelint
		b.Run(c.Name, func(b *testing.B) {
			// Read input.
			src, err := ioutil.ReadFile(c.Input)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = Format(src)
			}
		})
	}
}
