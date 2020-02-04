.PHONY: lint
lint:
	golangci-lint run

.PHONY: bootstrap
bootstrap:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b ${GOPATH}/bin v1.19.1

unimathsymbols.txt:
	wget http://milde.users.sourceforge.net/LUCR/Math/data/unimathsymbols.txt
