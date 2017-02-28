# compile depedencies and compile all tests
set -x 
mkdir -p build/tests
pushd build/tests
for i in `go list ../../integration_tests/...`; do go test -i $i; done
for i in `go list ../../integration_tests/...`; do go test -c $i; done
popd
