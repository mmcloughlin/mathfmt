.PHONY: lint
lint:
	golangci-lint run
	golangci-lint run make_symbols.go

.PHONY: generate
generate: symbols
	embedmd -w README.md

.PHONY: symbols
symbols: zsymbols.go symbols.md

zsymbols.go: make_symbols.go unimathsymbols.txt
	go run $< -input unimathsymbols.txt -type table -output $@

symbols.md: make_symbols.go unimathsymbols.txt
	go run $< -input unimathsymbols.txt -type doc -output $@

unimathsymbols.txt:
	wget http://milde.users.sourceforge.net/LUCR/Math/data/unimathsymbols.txt

.PHONY: clean
clean:
	$(RM) unimathsymbols.txt

.PHONY: bootstrap
bootstrap:
	go install github.com/campoy/embedmd@v1.0.0
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/v1.54.2/install.sh | sh -s -- -b ${GOPATH}/bin v1.54.2
