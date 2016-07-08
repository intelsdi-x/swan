package kubernetes

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/kubernetes"
	. "github.com/smartystreets/goconvey/convey"
	"time"
)

// Please see `pkg/kubernetes/README.md` for prerequisites for this test.
func TestLocalKubernetesPodExecution(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	Convey("While having local executor", t, func() {
		local := executor.NewLocal()
		Convey("We are able to launch kubernetes cluster on one node", func() {
			k8sLauncher := kubernetes.New(local, local, kubernetes.DefaultConfig())
			k8sHandle, err := k8sLauncher.Launch()
			So(err, ShouldBeNil)

			defer func() {
				err := k8sHandle.Stop()
				So(err, ShouldBeNil)
				err = k8sHandle.Clean()
				So(err, ShouldBeNil)
				err = k8sHandle.EraseOutput()
				So(err, ShouldBeNil)
			}()

			Convey("And kubectl is able to list the pods", func() {
				k8sHandle.Wait(0 * time.Nanosecond)
			})
		})
	})
}
