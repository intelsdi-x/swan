.PHONY: build

# Place for custom options for test commands.
TEST_OPT?=

# for compatibility purposes
# in the future deps target should point to deps_all, currently Kopernik job
# is running deps before running integration_tests. This is not needed, because
# we are downloading all of dependancies in provision phase.
deps: show_env
integration_test: build_plugins build_swan test_integration
unit_test: deps_godeps test_unit

deps_all: deps_godeps deps_other
build_all: deps_all build_workloads build_plugins build_swan
build_and_test_integration: build_all test_integration
build_and_test_unit: build_all test_lint test_unit
build_and_test_all: build_all test_lint test_unit test_integration
test_all: test_lint test_unit test_integration


# deps not covered by "vendor" folder (testing/developing env) rather than application (excluding convey)
deps_godeps:
	go get github.com/golang/lint/golint
	go get github.com/GeertJohan/fgt # return exit, fgt runs any command for you and exits with exitcode 1
	go get github.com/stretchr/testify
	go get github.com/vektra/mockery/.../
	go get github.com/Masterminds/glide
	glide install

deps_other:
	# Some workloads are Git Submodules
	git submodule update --init --recursive
	# Jupyter building
	(cd scripts/jupyter; sudo pip install -r requirements.txt)
	# Get SPECjbb
	(sudo bash scripts/get_specjbb.sh)

build_plugins:
	mkdir -p build
	(go get github.com/intelsdi-x/snap-plugin-publisher-cassandra)
	(go get github.com/intelsdi-x/snap-plugin-processor-tag)
	(go get github.com/intelsdi-x/kubesnap-plugin-collector-docker)
	(cd $(GOPATH)/src/github.com/intelsdi-x/kubesnap-plugin-collector-docker && patch -p1 --forward -s --merge < ../swan/misc/kubesnap_docker_collector.patch)
	(go install github.com/intelsdi-x/kubesnap-plugin-collector-docker)
	(go install ./misc/snap-plugin-collector-session-test)
	(go install ./misc/snap-plugin-publisher-session-test)
	(go install ./misc/snap-plugin-collector-mutilate)

build_workloads:
	(cd workloads/data_caching/memcached && ./build.sh)
	(cd workloads/low-level-aggressors && make -j4)

	# Prepare & Build Caffe workload.
	(cd ./workloads/deep_learning/caffe && cp caffe_cpu_solver.patch ./caffe_src/)
	(cd ./workloads/deep_learning/caffe && cp vagrant_vboxsf_workaround.patch ./caffe_src/)
	(cd ./workloads/deep_learning/caffe && cp get_cifar10.patch ./caffe_src/)
	(cd ./workloads/deep_learning/caffe/caffe_src/ && patch -p1 --forward -s --merge < caffe_cpu_solver.patch)
	(cd ./workloads/deep_learning/caffe/caffe_src/ && patch -p1 --forward -s --merge < vagrant_vboxsf_workaround.patch)
	(cd ./workloads/deep_learning/caffe/caffe_src/ && patch -p1 --forward -s --merge < get_cifar10.patch)
	(cd ./workloads/deep_learning/caffe && cp Makefile.config ./caffe_src/)
	(cd ./workloads/deep_learning/caffe/caffe_src && make -j4 all)
	(cd ./workloads/deep_learning/caffe && ./prepare_cifar10_dataset.sh)

build_swan:
	mkdir -p build/experiments/memcached
	(cd build/experiments/memcached; go build ../../../experiments/memcached-sensitivity-profile)

# testing
## fgt: lint doesn't return exit code when finds something (https://github.com/golang/lint/issues/65)
test_lint:
	fgt golint ./pkg/...
	fgt golint ./experiments/...
	fgt golint ./misc/...
	fgt golint ./integration_tests/...

test_unit:
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./pkg/... -v
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./experiments/... -v
	./scripts/isolate-pid.sh go test $(TEST_OPT) ./misc/... -v

test_integration:
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

repository_reset: cleanup
	(cd workloads/deep_learning/caffe/caffe_src/; git clean -fddx; git reset --hard)

show_env:
	@ echo Environment variables:
	@ echo ""
	@ env
	@ echo ""
