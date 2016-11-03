.PHONY: build

# Place for custom options for test commands.
TEST_OPT?=

all: deps lint unit_test

deps:
	curl https://glide.sh/get | /bin/bash
	go get github.com/golang/lint/golint
	go get github.com/GeertJohan/fgt
	glide install

# testing
## fgt: lint doesn't return exit code when finds something (https://github.com/golang/lint/issues/65)
lint:
	fgt golint ./pkg/...
	fgt golint ./integration_tests/...

unit_test:
	go test -race $(TEST_OPT) ./pkg/...

integration_test:
	go test -i -tags 'parallel sequential' ./integration_tests/...
	./scripts/isolate-pid.sh go test -race $(TEST_OPT) -tags 'parallel' ./integration_tests/...
	./scripts/isolate-pid.sh go test -race $(TEST_OPT) -tags 'sequential' -p=1 ./integration_tests/...
    # Run tests without tags in case of untagged test slip into repository.
	./scripts/isolate-pid.sh go test -race $(TEST_OPT) ./integration_tests/...
