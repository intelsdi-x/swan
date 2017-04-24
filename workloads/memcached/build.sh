set -e
mkdir -p output
docker build -t memcached .
docker run --rm memcached >output/memcached
chmod +x output/memcached
./output/memcached -V
