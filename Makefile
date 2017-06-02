PACKAGES=$(shell go list ./... | grep -v '/vendor/')

all: get_vendor_deps test install

test: test_lightcli test_nodecli test_unit

test_lightcli: 
	wget "https://raw.githubusercontent.com/kward/shunit2/master/source/2.1/src/shunit2" \
		-q -O test/shunit2 
	bash test/lightcli.sh

test_nodecli:
	wget "https://raw.githubusercontent.com/kward/shunit2/master/source/2.1/src/shunit2" \
		-q -O test/shunit2 
	bash test/nodecli.sh

test_unit:
	@go test $(PACKAGES)

install:
	@go install ./cmd/...

get_vendor_deps:
	go get github.com/Masterminds/glide
	glide install

.PHONY: install test test_unit test_cli get_vendor_deps 
