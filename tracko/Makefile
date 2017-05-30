PACKAGES=$(shell go list ./... | grep -v '/vendor/')

all: get_vendor_deps test install

test: test_cli test_unit

test_cli:
	wget "https://raw.githubusercontent.com/kward/shunit2/master/source/2.1/src/shunit2" \
		-q -O cmd/tracko/test/shunit2 
	bash cmd/tracko/test/test.sh

test_unit:
	@go test $(PACKAGES)

install:
	@go install ./cmd/...

get_vendor_deps:
	go get github.com/Masterminds/glide
	glide install

.PHONY: install test test_unit test_cli get_vendor_deps 
