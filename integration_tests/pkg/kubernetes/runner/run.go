package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/kubernetes"
	"time"
)

func check(err error) {
	if err != nil {
		logrus.Debugf("%+v", err)
		logrus.Fatalf("%v", err)
	}
}

// Please see `pkg/kubernetes/README.md` for prerequisites for this test.
func main() {
	logrus.SetLevel(logrus.DebugLevel)

	conf.ParseFlags()

	local := executor.NewLocal()
	k8sLauncher := kubernetes.New(local, local, kubernetes.DefaultConfig())
	k8sHandle, err := k8sLauncher.Launch()
	check(err)

	defer func() {
		k8sHandle.Stop()
		k8sHandle.Clean()
		k8sHandle.EraseOutput()
	}()

	k8sHandle.Wait(0 * time.Millisecond)
}
