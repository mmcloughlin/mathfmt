# mathfmt

Document mathematical Go code beautifully with Unicode symbols.

## Install

```
go get github.com/mmcloughlin/mathfmt
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

Run it through `mathfmt -w` and voila!

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

## License

`mathfmt` is available under the [BSD 3-Clause License](LICENSE).
