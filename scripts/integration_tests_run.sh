# run all tests as bins
set -x -e
for i in `ls build/tests/`; do ./build/tests/$i; done
