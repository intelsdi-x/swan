.PHONY: build

# Place for custom options for test commands.
TEST_OPT?=""

# for compatibility purposes.
deps: deps_all
integration_test: show_env build_plugins build_swan test_integration
unit_test: deps_godeps test_unit

deps_all: deps_godeps deps_jupyter
build_all: deps_all build_workloads build_plugins build_swan
build_and_test_integration: build_all test_integration
build_and_test_unit: build_all test_lint test_unit
build_and_test_all: build_all test_lint test_unit test_integration
test_all: test_lint test_unit test_integration


# deps not covered by "vendor" folder (testing/developing env) rather than application (excluding convey)
deps_godeps:
	go get -u github.com/golang/lint/golint
	go get -u github.com/GeertJohan/fgt # return exit, fgt runs any command for you and exits with exitcode 1
	go get github.com/stretchr/testify # go get -u github.com/stretchr/testify fails miserably
	go get -u github.com/vektra/mockery/.../
	go get -u github.com/Masterminds/glide
	glide install

deps_jupyter:
	# Jupyter building
	(cd scripts/jupyter; sudo pip install -r requirements.txt)

build_plugins:
	(./scripts/build_plugins.sh)

build_image:
	(./scripts/build_docker_image.sh)

build_workloads:
	# Some workloads are Git Submodules
	git submodule update --init --recursive

	(cd workloads/data_caching/memcached && ./build.sh)
	(cd workloads/low-level-aggressors && make -j4)

	# Prepare & Build Caffe workload.
	(cd ./workloads/deep_learning/caffe && ./build_caffe.sh ${BUILD_OPENBLAS})

	# Get SPECjbb
	(sudo ./scripts/get_specjbb.sh)

build_swan:
	mkdir -p build/experiments/memcached build/experiments/specjbb
	(cd build/experiments/memcached; go build ../../../experiments/memcached-sensitivity-profile)
	(cd build/experiments/specjbb; go build ../../../experiments/specjbb-sensitivity-profile)

dist: build_workloads build_swan
	(./scripts/artifacts.sh dist)

install:
	(./scripts/artifacts.sh install)

uninstall:
	(./scripts/artifacts.sh uninstall)

# testing
## fgt: lint doesn't return exit code when finds something (https://github.com/golang/lint/issues/65)
test_lint:
	fgt golint ./pkg/...
	fgt golint ./experiments/...
	fgt golint ./misc/...
	fgt golint ./integration_tests/...

test_unit:
	./scripts/isolate-pid.sh go test $(TEST_OPT) -v ./pkg/...
	./scripts/isolate-pid.sh go test $(TEST_OPT) -v ./experiments/...
	./scripts/isolate-pid.sh go test $(TEST_OPT) -v ./misc/...

test_integration:
	go test -i ./integration_tests/... ./experiments/... ./misc/...
	./scripts/isolate-pid.sh go test -p 1 -v $(TEST_OPT) ./integration_tests/...
	./scripts/isolate-pid.sh go test -p 1 -v $(TEST_OPT) ./experiments/...
	./scripts/isolate-pid.sh go test -p 1 -v $(TEST_OPT) ./misc/...
	(cd scripts/jupyter; py.test)

cleanup:
	rm -fr misc/**/*log
	rm -fr integration_tests/**/*log
	rm -fr integration_tests/**/remote_memcached_*
	rm -fr integration_tests/**/local_snapd_*

remove_vendor:
	rm -fr vendor/

repository_reset: cleanup remove_vendor
	(cd workloads/deep_learning/caffe/caffe_src/; git clean -fddx; git reset --hard)
	(cd workloads/deep_learning/caffe/openblas/; git clean -fddx; git reset --hard)

show_env:
	@ echo Environment variables:
	@ echo ""
	@ env
	@ echo ""
