echo "Getting golint if not found"
golint || go get -u github.com/golang/lint/golint

echo "Checking dummy style"
sh -c "cd pkg/dummy && golint"

echo "Checking shell style"
sh -c "cd pkg/shell && golint"
