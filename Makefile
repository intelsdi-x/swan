# Copyright (c) 2017 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: build

# Place for custom options for test commands.
TEST_OPT?=

# High level targets called from travis with dependencies.
integration_test: show_env restart_snap deps build dist install test_integration
unit_test: deps test_integration_build build test_unit 
lint: linter test_lint

build: build_swan build_plugins
build_all: deps build_plugins build_swan
build_and_test_integration: build_all test_integration test_jupyter_unit
build_and_test_unit: build_all test_lint test_unit
build_and_test_all: build_all test_all
test_all: test_lint test_unit test_integration test_jupyter_unit

restart_snap:
	# Workaround for "Snap does not refresh hostname" https://github.com/intelsdi-x/snap/issues/1514
	sudo systemctl restart snap-telemetry

glide:
	curl https://glide.sh/get | sh
	rm -fr vendor
	glide -q install

linter:
	go get github.com/alecthomas/gometalinter
	gometalinter --install
	
deps: glide linter
	go get github.com/stretchr/testify

build_plugins:
	mkdir -p build/plugins
	(cd build/plugins; go build ../../plugins/snap-plugin-publisher-session-test)
	(cd build/plugins; go build ../../plugins/snap-plugin-collector-mutilate)
	(cd build/plugins; go build ../../plugins/snap-plugin-collector-specjbb)
	(cd build/plugins; go build ../../plugins/snap-plugin-collector-caffe-inference)

build_swan:
	go build -i -v ./experiments/...
	mkdir -p build/experiments/memcached build/experiments/specjbb build/experiments/optimal-core-allocation build/experiments/memcached-cat build/experiments/example
	(cd build/experiments/memcached; go build ../../../experiments/memcached-sensitivity-profile)
	(cd build/experiments/specjbb; go build ../../../experiments/specjbb-sensitivity-profile)
	(cd build/experiments/optimal-core-allocation; go build ../../../experiments/optimal-core-allocation)
	(cd build/experiments/memcached-cat; go build ../../../experiments/memcached-cat)
	(cd build/experiments/example; go build ../../../experiments/example)

# testing
test_lint:
	GOMAXPROCS=2 gometalinter --config=.lint ./pkg/...
	GOMAXPROCS=2 gometalinter --config=.lint ./experiments/...
	GOMAXPROCS=2 gometalinter --config=.lint ./plugins/...
	GOMAXPROCS=2 gometalinter --config=.lint ./integration_tests/...

test_jupyter_lint: jupyter_image
	docker run --rm intelsdi/swan-jupyter pycodestyle --max-line-length=120 .

test_jupyter_unit: jupyter_image
	docker run --rm intelsdi/swan-jupyter python test_swan.py

test_unit:
	go test -i ./pkg/... ./plugins/...
	go test -p 1 $(TEST_OPT) ./pkg/... ./plugins/...

# make sure that all integration tests are building without problem - not required directly for test_integration (only used by .travis)
test_integration_build:
	./scripts/integration_tests_build.sh

test_integration:
	go test -i ./integration_tests/... 
	@# Validate that there is enough cpus to run integration tests.
	@if [[ `nproc` -lt 2 ]]; then \
		echo 'warning: not enough cpus `nproc` to run integrations tests (two cores are required)'; \
		exit 1;\
	fi
	./scripts/isolate-pid.sh go test -p 1 $(TEST_OPT) ./integration_tests
	./scripts/isolate-pid.sh go test -p 1 $(TEST_OPT) ./integration_tests/pkg/...
	./scripts/isolate-pid.sh go test -p 1 $(TEST_OPT) ./integration_tests/snap-plugins/...
	./scripts/isolate-pid.sh go test -p 1 $(TEST_OPT) ./integration_tests/experiments/...

cleanup:
	rm -fr plugins/**/*log
	rm -fr integration_tests/**/*log
	rm -fr integration_tests/**/remote_memcached_*
	rm -fr integration_tests/**/local_snapteld_*

remove_vendor:
	rm -fr vendor/

show_env:
	@ echo Environment variables:
	@ echo ""
	@ env
	@ echo ""

dist:
	tar -C ./build/experiments/memcached -cvf swan.tar memcached-sensitivity-profile 
	tar -C ./build/experiments/specjbb -rvf swan.tar specjbb-sensitivity-profile
	tar -C ./build/experiments/optimal-core-allocation -rvf swan.tar optimal-core-allocation
	tar -C ./build/experiments/memcached-cat -rvf swan.tar memcached-cat
	tar -C ./build/experiments/example -rvf swan.tar example
	tar -C ./build/plugins -rvf swan.tar snap-plugin-collector-caffe-inference snap-plugin-collector-mutilate snap-plugin-collector-specjbb snap-plugin-publisher-session-test
	tar --transform 's/-binary//' -rvf swan.tar NOTICE-binary
	tar -rvf swan.tar LICENSE
	gzip -f swan.tar

install:
	tar -C /opt/swan/bin -xzvf swan.tar.gz 
	sudo ln -svf /opt/swan/bin/* /usr/bin/

docker:
	docker build -t intelsdi/swan:latest workloads

extract_binaries:
	docker run --rm -v $(PWD)/opt:/output intelsdi/swan cp -R /opt/swan /output

jupyter_image:
	docker build -t intelsdi/swan-jupyter jupyter
