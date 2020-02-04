.PHONY: symbols
symbols: zsymbols.go symbols.md

zsymbols.go: make_symbols.go unimathsymbols.txt
	go run $< -input unimathsymbols.txt -type table -output $@

symbols.md: make_symbols.go unimathsymbols.txt
	go run $< -input unimathsymbols.txt -type doc -output $@

unimathsymbols.txt:
	wget http://milde.users.sourceforge.net/LUCR/Math/data/unimathsymbols.txt

.PHONY: lint
lint:
	golangci-lint run
	golangci-lint run make_symbols.go

.PHONY: bootstrap
bootstrap:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b ${GOPATH}/bin v1.19.1
