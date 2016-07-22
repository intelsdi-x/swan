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
	# TODO(skonefal): Remove go build when we will load plugins from single place in code.
	(cd build; go build ../misc/snap-plugin-collector-session-test)
	(cd build; go build ../misc/snap-plugin-publisher-session-test)
	(cd build; go build ../misc/snap-plugin-collector-mutilate)
	(go get github.com/intelsdi-x/snap-plugin-publisher-cassandra)
	(go install github.com/intelsdi-x/snap-plugin-publisher-cassandra)
	(go install github.com/intelsdi-x/snap-plugin-processor-tag)
	(go install ./misc/snap-plugin-collector-session-test)
	(go install ./misc/snap-plugin-publisher-session-test)
	(go install ./misc/snap-plugin-collector-mutilate)

list_env:
	@ echo Environment variables:
	@ echo ""
	@ env
	@ echo ""

integration_test: list_env plugins unit_test build_workloads build build_jupyter
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./integration_tests/... -v
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./experiments/...
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./misc/...

# For development purposes we need to do integration test faster on already built components.
integration_test_no_build: list_env unit_test build
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./integration_tests/...
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./experiments/...

integration_test_on_docker:
	(cd integration_tests/docker; ./inside-docker-tests.sh)

# building
build:
	mkdir -p build/experiments/memcached
	(cd build/experiments/memcached; go build ../../../experiments/memcached-sensitivity-profile)

build_jupyter:
	(cd scripts/jupyter; sudo pip install -r requirements.txt)
	(cd scripts/jupyter; py.test)

build_workloads:
	(cd workloads/data_caching/memcached && ./build.sh)
	(cd workloads/low-level-aggressors && make -j4)

	# Some workloads are Git Submodules
	git submodule update --init --recursive

	# Prepare & Build Caffe workload.
	(cd ./workloads/deep_learning/caffe && cp caffe_cpu_solver.patch ./caffe_src/)
	(cd ./workloads/deep_learning/caffe && cp vagrant_vboxsf_workaround.patch ./caffe_src/)
	(cd ./workloads/deep_learning/caffe/caffe_src/ && patch -p1 --forward -s --merge < caffe_cpu_solver.patch)
	(cd ./workloads/deep_learning/caffe/caffe_src/ && patch -p1 --forward -s --merge < vagrant_vboxsf_workaround.patch)
	(cd ./workloads/deep_learning/caffe && cp Makefile.config ./caffe_src/)
	(cd ./workloads/deep_learning/caffe/caffe_src && make -j4 all)
	(cd ./workloads/deep_learning/caffe && ./prepare_cifar10_dataset.sh)
	# Install kubernetes binaries.
	(bash ./misc/kubernetes/install_binaries.sh)

cleanup:
	rm -fr misc/**/*log
	rm -fr integration_tests/**/*log
	rm -fr integration_tests/**/remote_memcached_*
	rm -fr integration_tests/**/local_snapd_*
