.PHONY: build

# Place for custom options for test commands.
TEST_OPT?=

# for compatibility purposes.
integration_test: show_env restart_snap deps build dist install test_integration
unit_test: deps test_unit

build: build_swan build_plugins
build_all: deps build_plugins build_swan
build_and_test_integration: build_all test_integration
build_and_test_unit: build_all test_lint test_unit
build_and_test_all: build_all test_all
test_all: test_lint test_unit test_unit_jupyter test_integration e2e_test

restart_snap:
	# Workaround for "Snap does not refresh hostname" https://github.com/intelsdi-x/snap/issues/1514
	sudo systemctl restart snap-telemetry

glide:
	# Workaround for https://github.com/Masterminds/glide/issues/784
	mkdir -p ${GOPATH}/bin
	wget -q https://github.com/Masterminds/glide/releases/download/v0.12.3/glide-v0.12.3-linux-386.tar.gz -O - | tar xzv --strip-components 1 -C ${GOPATH}/bin linux-386/glide
	curl -s https://glide.sh/get | sh
	glide -q install
	
deps: glide
	# Warning: do not try to update (-u) because it fails (upstream changed in no updateable manner).
	go get github.com/alecthomas/gometalinter
	go get github.com/stretchr/testify
	gometalinter --install

build_plugins:
	mkdir -p build/plugins
	(cd build/plugins; go build ../../plugins/snap-plugin-publisher-session-test)
	(cd build/plugins; go build ../../plugins/snap-plugin-collector-mutilate)
	(cd build/plugins; go build ../../plugins/snap-plugin-collector-specjbb)
	(cd build/plugins; go build ../../plugins/snap-plugin-collector-caffe-inference)

build_swan:
	mkdir -p build/experiments/memcached build/experiments/specjbb
	(cd build/experiments/memcached; go build ../../../experiments/memcached-sensitivity-profile)
	(cd build/experiments/specjbb; go build ../../../experiments/specjbb-sensitivity-profile)


# testing
test_lint:
	gometalinter --config=.lint ./pkg/...
	gometalinter --config=.lint ./experiments/...
	gometalinter --config=.lint ./plugins/...
	gometalinter --config=.lint ./integration_tests/...
	pep8 --max-line-length=120 jupyter/

test_unit:
	go test -i ./pkg/... ./plugins/...
	go test -p 1 $(TEST_OPT) ./pkg/... ./plugins/...

# make sure that all integration tests are building without problem - not required directly for test_integration (only used by .travis)
test_integration_build:
	./scripts/integration_tests_build.sh

test_integration:
	go test -i ./integration_tests/... 
	./scripts/isolate-pid.sh go test -p 1 $(TEST_OPT) ./integration_tests/... 

deps_jupyter:
	# Required for jupyter building.
	sudo yum install -y gcc
	(cd jupyter; sudo pip install -r requirements.txt)

e2e_test: deps_jupyter
	(cd jupyter; py.test)
	sudo service snapteld start
	SWAN_LOG=debug SWAN_BE_SETS=0:0 SWAN_HP_SETS=0:0 sudo -E memcached-sensitivity-profile --aggr caffe > jupyter/integration_tests/experiment_id.stdout
	sudo service snapteld stop
	jupyter nbconvert --execute --to notebook --inplace jupyter/integration_tests/integration_tests.ipynb
	rm jupyter/integration_tests/integration_tests.py jupyter/integration_tests/*.stdout

cleanup:
	rm -fr plugins/**/*log
	rm -fr integration_tests/**/*log
	rm -fr integration_tests/**/remote_memcached_*
	rm -fr integration_tests/**/local_snapteld_*
	rm -fr jupyter/integration_tests/*.stdout

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
	tar -C ./build/plugins -rvf swan.tar snap-plugin-collector-caffe-inference snap-plugin-collector-mutilate snap-plugin-collector-specjbb snap-plugin-publisher-session-test
	gzip -f swan.tar

install:
	tar -C /opt/swan/bin -xzvf swan.tar.gz 
	sudo ln -svf /opt/swan/bin/* /usr/bin/
