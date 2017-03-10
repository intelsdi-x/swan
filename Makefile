.PHONY: build

# Place for custom options for test commands.
TEST_OPT?=

# for compatibility purposes.
deps: deps_godeps
integration_test: show_env deps_godeps build_plugins build_swan test_integration_build test_integration
unit_test: deps test_unit

build_all: deps_godeps build_workloads build_plugins build_swan
build_and_test_integration: build_all test_integration
build_and_test_unit: build_all test_lint test_unit
build_and_test_all: build_all test_all
test_all: test_lint test_unit test_unit_jupyter test_integration e2e_test


# deps not covered by "vendor" folder (testing/developing env) rather than application 
deps_godeps:
	go get github.com/golang/lint/golint
	go get github.com/GeertJohan/fgt 
	go get github.com/stretchr/testify 
	# only required for generating mocks.
	#go get github.com/vektra/mockery/...
	curl -s https://glide.sh/get | sh
	glide install

build_plugins:
	(go install ./misc/snap-plugin-publisher-session-test)
	(go install ./misc/snap-plugin-collector-mutilate)
	(go install ./misc/snap-plugin-collector-specjbb)
	(go install ./misc/snap-plugin-collector-caffe-inference)

build_swan:
	mkdir -p build/experiments/memcached build/experiments/specjbb
	(cd build/experiments/memcached; go build ../../../experiments/memcached-sensitivity-profile)
	(cd build/experiments/specjbb; go build ../../../experiments/specjbb-sensitivity-profile)

# testing
## fgt: lint doesn't return exit code when finds something (https://github.com/golang/lint/issues/65)
test_lint:
	fgt golint ./pkg/...
	fgt golint ./experiments/...
	fgt golint ./misc/...
	fgt golint ./integration_tests/...

test_unit:
	go test -i ./pkg/... ./misc/...
	go test -p 1 $(TEST_OPT) ./pkg/... ./misc/...

# make sure that all integration tests are building without problem - not required directly for test_integration
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
	rm -fr misc/**/*log
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
