mkdir output
docker build -t stress-ng .
docker run -v $PWD/output:/output --rm stress-ng cp -v stress-ng /output/
