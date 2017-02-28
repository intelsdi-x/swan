# memcached workload

## Install dependencies

From repo:

```
$ sudo yum install autoconf scons gengetopt automake libevent-devel
```

Bundled binaries:

```
$ sudo rpm -i dependencies/scons-2.4.1-1.noarch.rpm
$ sudo rpm -i dependencies/gengetopt-2.22.6-1.el7.x86_64.rpm
```

## Build memcached and mutilate

```
$ ./build.sh
```
