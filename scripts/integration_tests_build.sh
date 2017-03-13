# compile depedencies and compile all tests
set -e 
echo "install dependecies..."
go test -i ./integration_tests/...
echo "build tests..."
mkdir -p build/tests
cd build/tests
for i in `go list ../../integration_tests/...`; do 
    echo building: $i
    go test -c $i
done
