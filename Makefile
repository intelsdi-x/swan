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
	fgt golint ./misc/...

unit_test:
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./pkg/...
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./experiments/...
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./misc/...

plugins:
	(cd build; go build ../misc/snap-plugin-collector-session-test)
	(cd build; go build ../misc/snap-plugin-processor-session-tagging)
	(cd build; go build ../misc/snap-plugin-publisher-session-test)
	(cd misc/snap-plugin-collector-mutilate; go build)

integration_test: plugins
	./scripts/isolate-pid.sh go test $(TEST_OPT) -tags=integration ./pkg/...
	./scripts/isolate-pid.sh go test $(TEST_OPT) -tags=integration ./experiments/...
#   TODO(niklas): Fix race (https://intelsdi.atlassian.net/browse/SCE-316)
#	./scripts/isolate-pid.sh go test $(TEST_OPT) -tags=integration ./misc/...
	./scripts/isolate-pid.sh go test -tags=integration ./misc/...

# building
build:
	mkdir -p build
	(cd build; go build ../experiments/...)

cleanup:
	rm -f misc/snap-plugin-collector-mutilate/????-??-??_snap-plugin-collector-mutilate.log
	rm -f misc/snap-plugin-collector-mutilate/????-??-??_snap-plugin-collector-mutilate.test.log
	rm -f misc/snap-plugin-collector-mutilate/mutilate/????-??-??_mutilate.test.log
