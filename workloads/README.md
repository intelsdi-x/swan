# Swan Docker image

The image allows to run follwing applications:
* memcached - first [High Priority workload](https://memcached.org) supported;
* stress-ng - set of [synthetic aggressors](https://github.com/ColinIanKing/stress-ng) that allow to analyze interference with memcached;
* caffe - [deep learning framework](http://caffe.berkeleyvision.org) with [Cifar-10](http://caffe.berkeleyvision.org/gathered/examples/cifar10.html) an example of real world Best Effort workload;
* intel-cmt-cat - set of [Intel open source tools](https://github.com/01org/intel-cmt-cat) that allow to manipulate low-level resource isolation.

## Fetching image from hub.docker.com

We recommend fetching the image from hub.docker.com by calling:

```sh
docker pull intelsdi/swan
```

## Building image on your own

To build image on your own you will need to use version 17.05 or newer of Docker. Image is being built using Community Edition. If you want to try then just run:

```sh
make docker
```

from project's main directory.
