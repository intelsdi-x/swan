.PHONY: build

# Place for custom options for test commands.
TEST_OPT?=

build_all: build_workloads build_plugins build_swan
build_and_test_integration: build_all test_integration
build_and_test_unit: deps_godeps test_lint test_unit
build_and_test_all: build_all test_lint test_unit test_integration
test_all: test_lint test_unit test_integration

# deps not covered by "vendor" folder (testing/developing env) rather than application (excluding convey)
deps_godeps:
	go get github.com/tools/godep
	go get github.com/golang/lint/golint
	go get github.com/GeertJohan/fgt # return exit, fgt runs any command for you and exits with exitcode 1
	go get github.com/stretchr/testify
	go get github.com/vektra/mockery/.../
	godep restore -v

deps_all: deps_godeps
	# Some workloads are Git Submodules
	git submodule update --init --recursive
	# Jupyter building
	(cd scripts/jupyter; sudo pip install -r requirements.txt)
	# Install kubernetes binaries.
	(bash ./misc/kubernetes/install_binaries.sh)

build_plugins: deps_godeps
	mkdir -p build
	(cd build; go build ../misc/snap-plugin-collector-session-test)
	(cd build; go build ../misc/snap-plugin-publisher-session-test)
	(cd build; go build ../misc/snap-plugin-collector-mutilate)
	(go get github.com/intelsdi-x/snap-plugin-publisher-cassandra)
	(go install github.com/intelsdi-x/snap-plugin-processor-tag)

build_workloads: deps_godeps
	(cd workloads/data_caching/memcached && ./build.sh)
	(cd workloads/low-level-aggressors && make -j4)

	# Prepare & Build Caffe workload.
	(cd ./workloads/deep_learning/caffe && cp caffe_cpu_solver.patch ./caffe_src/)
	(cd ./workloads/deep_learning/caffe && cp vagrant_vboxsf_workaround.patch ./caffe_src/)
	(cd ./workloads/deep_learning/caffe/caffe_src/ && patch -p1 --forward -s --merge < caffe_cpu_solver.patch)
	(cd ./workloads/deep_learning/caffe/caffe_src/ && patch -p1 --forward -s --merge < vagrant_vboxsf_workaround.patch)
	(cd ./workloads/deep_learning/caffe && cp Makefile.config ./caffe_src/)
	(cd ./workloads/deep_learning/caffe/caffe_src && make -j4 all)
	(cd ./workloads/deep_learning/caffe && ./prepare_cifar10_dataset.sh)

build_swan: deps_all
	mkdir -p build/experiments/memcached
	(cd build/experiments/memcached; go build ../../../experiments/memcached-sensitivity-profile)

# testing
## fgt: lint doesn't return exit code when finds something (https://github.com/golang/lint/issues/65)
test_lint:
	fgt golint ./pkg/...
	fgt golint ./experiments/...
	fgt golint ./misc/...
	fgt golint ./integration_tests/...
	fgt golint ./scripts/...

test_unit: deps_godeps
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./pkg/... -v
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./experiments/... -v
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./misc/... -v

test_integration: build_swan
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./integration_tests/... -v
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./experiments/... -v
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./misc/... -v
	(cd scripts/jupyter; py.test)

test_integration_on_docker:
	(cd integration_tests/docker; ./inside-docker-tests.sh)	

cleanup:
	rm -fr misc/**/*log
	rm -fr integration_tests/**/*log
	rm -fr integration_tests/**/remote_memcached_*
	rm -fr integration_tests/**/local_snapd_*

show_env:
	@ echo Environment variables:
	@ echo ""
	@ env
	@ echo ""
