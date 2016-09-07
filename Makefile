.PHONY: build

# Place for custom options for test commands.
TEST_OPT?=

all: deps lint unit_test

deps:
	curl https://glide.sh/get | sh
	go get github.com/golang/lint/golint
	go get github.com/GeertJohan/fgt
	glide install

# testing
## fgt: lint doesn't return exit code when finds something (https://github.com/golang/lint/issues/65)
lint:
	fgt golint ./pkg/...
	fgt golint ./integration_tests/...

unit_test:
	go test $(TEST_OPT) ./pkg/...

integration_test:
	go test $(TEST_OPT) ./integration_tests/...
