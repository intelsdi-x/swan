.PHONY: build

# Place for custom options for test commands.
TEST_OPT?=

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
	fgt golint ./integration_tests/...

unit_test:
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./pkg/...
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./experiments/...
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./misc/...

plugins:
	mkdir -p build
	(cd build; go build ../misc/snap-plugin-collector-session-test)
	(cd build; go build ../misc/snap-plugin-publisher-session-test)
	(cd build; go build ../misc/snap-plugin-collector-mutilate)

integration_test: plugins unit_test build_workloads
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./integration_tests/...
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./experiments/...
#   TODO(niklas): Fix race (https://intelsdi.atlassian.net/browse/SCE-316)
#	./scripts/isolate-pid.sh go test $(TEST_OPT) ./misc/...

# For development purposes we need to do integration test faster on already built components.
integration_test_no_build: unit_test
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./integration_tests/...
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./experiments/...

# building
build:
	mkdir -p build
	(cd build; go build ../experiments/...)

build_workloads:
	(cd workloads/data_caching/memcached; ./build.sh)
	(cd workloads/low-level-aggressors; make)

cleanup:
	rm -f misc/snap-plugin-collector-mutilate/????-??-??_snap-plugin-collector-mutilate.log
	rm -f misc/snap-plugin-collector-mutilate/????-??-??_snap-plugin-collector-mutilate.test.log
	rm -f misc/snap-plugin-collector-mutilate/mutilate/????-??-??_mutilate.test.log
	rm -rf integration_tests/pkg/executor/remote_memcached_*
	rm -fr integration_tests/pkg/snap/local_snapd_*
