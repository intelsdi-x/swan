if [ ! -d opt/swan ]; then
    echo "error: build opt/swan first: use 1_build_opt_swan.sh"
    exit 1
fi

echo building image
docker build -t centos_swan_image .

echo tests
docker run --rm centos_swan_image caffe.sh --version
docker run --rm centos_swan_image memcached -V

echo exporting to centos_swan_image.tgz
docker image save centos_swan_image -o centos_swan_image.tar
gzip centos_swan_image.tar
mv centos_swan_image.tar.gz centos_swan_image.tgz

echo to import docker image:
echo "cat centos_swan_image.tgz|gunzip|docker image load"
echo "docker images centos_swan_image"
