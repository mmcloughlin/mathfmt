# mathfmt

Document mathematical Go code beautifully with Unicode symbols.

* Write mathematical formulae in a [LaTeX](https://en.wikipedia.org/wiki/LaTeX)-ish syntax
* Super/subscripts formatted with Unicode characters: `2^32` becomes `2³²` and `x_{i+1}` becomes `xᵢ₊₁`
* Comprehensive [symbol library](symbols.md): `\zeta(s) = \sum 1/n^{s}` becomes `ζ(s) = ∑ 1/nˢ`

Inspired by [Filippo Valsorda](https://filippo.io/)'s [literate Go
implementation of
Poly1305](https://blog.filippo.io/a-literate-go-implementation-of-poly1305/).

## Usage

Install `mathfmt` with:

```
go get -u github.com/mmcloughlin/mathfmt
```

Apply to files just like you would with `gofmt`.

```
mathfmt -w file.go
```

## Example

Here's our variance function in Go, documented with LaTeX-ish equations in comments.

[embedmd]:# (testdata/stats.in go /\/\/ Variance/ /^}/)
```go
// Variance computes the population variance of the population x_{i} of size N.
// Specifically, it computes \sigma^2 where
//
//		\sigma^2 = \sum (x_{i} - \mu)^2 / N
//
// See also: https://en.wikipedia.org/wiki/Variance.
func Variance(X []float64) float64 {
	// Compute the average \mu.
	mu := Mean(X)

	// Compute the sum \sum (x_{i} - \mu)^2.
	ss := 0.0
	for _, x := range X {
		ss += (x - mu) * (x - mu) // (x_{i} - \mu)^2
	}

	// Final divide by N to produce \sigma^2.
	return ss / float64(len(X))
}
```

Run it through `mathfmt` and voila!

[embedmd]:# (testdata/stats.golden go /\/\/ Variance/ /^}/)
```go
// Variance computes the population variance of the population xᵢ of size N.
// Specifically, it computes σ² where
//
//		σ² = ∑ (xᵢ - μ)² / N
//
// See also: https://en.wikipedia.org/wiki/Variance.
func Variance(X []float64) float64 {
	// Compute the average μ.
	mu := Mean(X)

	// Compute the sum ∑ (xᵢ - μ)².
	ss := 0.0
	for _, x := range X {
		ss += (x - mu) * (x - mu) // (xᵢ - μ)²
	}

	// Final divide by N to produce σ².
	return ss / float64(len(X))
}
```

## Syntax

First a warning: `mathfmt` does not have a rigorous grammar, it's a
combination of string replacement and regular expressions that appears to
work most of time. However you may run into some [thorny edge
cases](https://github.com/mmcloughlin/mathfmt/issues/9).

* **Source:** `mathfmt` only works on Go source code. _Every_ comment in the file is
    processed, both single- and multi-line.
* **Symbols:** `mathfmt` recognizes a [huge symbol table](symbols.md) that is almost
    entirely borrowed from LaTeX packages. Every symbol macro in comment text
    will be replaced with its corresponding Unicode character. In addition to
    LaTeX symbol macros, `mathfmt` supports a [limited set of
    "aliases"](symbols.md#aliases) for character combinations commonly used to
    represent mathematical symbols.
* **Super/subscripts:** like LaTeX, superscripts use the `^` character and
    subscripts use `_`. If the super/subscript consists entirely of digits, then
    no braces are required: for example `2^128` or `x_13`. Otherwise braces must
    be used to surround the super/subscript, for example `2^{i}` or `x_{i+j}`.
    Note that Unicode support for super/subscripts is limited, and in particular
    does not support the full alphabet. Therefore, if there is not a
    corresponding super/subscript character available for any character in braces
    `{...}`, `mathfmt` will not perform any substition at all. For example there
    is no superscript `q`, so `mathfmt` will not be able to process `2^{q}`, and
    likewise with `x_{K}`.

## Credits

Thank you to Günter Milde for the exhaustive [`unimathsymbols`
database](http://milde.users.sourceforge.net/LUCR/Math/) of Unicode symbols
and corresponding LaTeX math mode commands.

## License

`mathfmt` is available under the [BSD 3-Clause License](LICENSE).
