.PHONY: build

# Place for custom options for test commands.
TEST_OPT?=-race

all: lint unit_test build cleanup

# deps not covered by "vendor" folder (testing/developing env) rather than application (excluding convey)
deps:
	go get github.com/tools/godep
	go get github.com/golang/lint/golint
	go get github.com/GeertJohan/fgt # return exit, fgt runs any command for you and exits with exitcode 1
	go get github.com/stretchr/testify
	go get github.com/vektra/mockery/.../
	godep restore -v

# testing
## fgt: lint doesn't return exit code when finds something (https://github.com/golang/lint/issues/65)
lint:
	fgt golint ./pkg/...
	fgt golint ./experiments/...
	fgt golint ./cmds/...
	fgt golint ./misc/...

unit_test:
	go test $(TEST_OPT) ./pkg/...
	go test $(TEST_OPT) ./experiments/...
	go test $(TEST_OPT) ./cmds/...
	go test $(TEST_OPT) ./misc/...

integration_test:
	go test $(TEST_OPT) -tags=integration ./pkg/...
	go test $(TEST_OPT) -tags=integration ./experiments/...
	go test $(TEST_OPT) -tags=integration ./cmds/...
	(cd misc/snap-plugin-collector-mutilate; go build; cd ../..)
	go test $(TEST_OPT) -tags=integration ./misc/...

# building
build:
	mkdir -p build
	(cd build; go build ../experiments/...)

cleanup:
	rm -f misc/snap-plugin-collector-mutilate/????-??-??_snap-plugin-collector-mutilate.log
	rm -f misc/snap-plugin-collector-mutilate/????-??-??_snap-plugin-collector-mutilate.test.log
	rm -f misc/snap-plugin-collector-mutilate/mutilate/????-??-??_mutilate.test.log
