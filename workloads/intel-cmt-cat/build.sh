mkdir output
docker build -t intel-cmt-cat .
docker run -v $PWD/output:/output --rm intel-cmt-cat cp -v pqos/pqos rdtset/rdtset /output/
