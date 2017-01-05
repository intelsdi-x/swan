## Building custom kube-apiserver

In order to add automatic assignment of high-priority taints and tolerations to guaranteed and
burstable pods, Athena provides an admission controller for the kube-apiserver.

In order to build the custom kube-apiserver, you need a golang environment (1.6 or newer) and `patch` and run:

```
go get -u github.com/jteeuwen/go-bindata/go-bindata
mkdir -p $GOPATH/src/k8s.io/kubernetes
git clone https://github.com/kubernetes/kubernetes.git $GOPATH/src/k8s.io/kubernetes
cd $GOPATH/src/k8s.io/kubernetes
git checkout v1.4.0-alpha.3
patch -p1 < $GOPATH/src/intelsdi-x/athena/misc/kubernetes/addtoleration.patch
make WHAT='cmd/kube-apiserver'
```

This should leave the kube-apiserver in `_output/bin/kube-apiserver`.
