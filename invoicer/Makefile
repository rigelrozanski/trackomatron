
PACKAGES=$(shell go list ./... | grep -v '/vendor/')

all: get_vendor_deps test install

test:
	bash cmd/invoicer/test/test.sh
	@go test $(PACKAGES)

install:
	@go install ./cmd/...

get_vendor_deps:
	go get github.com/Masterminds/glide
	glide install


.PHONY: install test get_vendor_deps
