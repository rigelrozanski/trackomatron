PACKAGES=$(shell go list ./... | grep -v '/vendor/')

all: get_vendor_deps install test

test: install test_unit test_cli

test_cli: 
	wget "https://raw.githubusercontent.com/kward/shunit2/master/source/2.1/src/shunit2" \
		-q -O test/shunit2 
	bash test/cli.sh

test_unit:
	@go test $(PACKAGES)

install:
	@go install ./cmd/...

get_vendor_deps:
	wget "https://raw.githubusercontent.com/LedgerHQ/ledger-wallet-api/master/ledger.js" \
		-q -O common/ledger.js
	go get github.com/Masterminds/glide
	glide install

.PHONY: install test test_unit test_cli get_vendor_deps 
