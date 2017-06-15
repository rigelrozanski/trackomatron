PACKAGES=$(shell go list ./... | grep -v '/vendor/')

all: get_vendor_deps install

test: install test_cli test_unit

test_cli: 
	wget "https://raw.githubusercontent.com/kward/shunit2/master/source/2.1/src/shunit2" \
		-q -O test/shunit2 
	bash test/cli.sh

test_unit:
	@go test $(PACKAGES)

install:
	@go install ./cmd/...

get_vendor_deps:
	go get github.com/Masterminds/glide
	glide install

.PHONY: install test test_unit test_lightcli test_heavycli get_vendor_deps 
