package kubernetes

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/kubernetes"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	. "github.com/smartystreets/goconvey/convey"
	"path"
	"testing"
	"time"
)

var (
	kubectlBinPath = path.Join(fs.GetSwanBinPath(), "kubectl")
)

func check(err error) {
	if err != nil {
		logrus.Debugf("%+v", err)
		logrus.Fatalf("%v", err)
	}
}

// Please see `pkg/kubernetes/README.md` for prerequisites for this test.
func TestLocalKubernetesRun(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)

	local := executor.NewLocal()
	k8sLauncher := kubernetes.New(local, local, kubernetes.DefaultConfig())
	k8sHandle, err := k8sLauncher.Launch()
	check(err)

	defer func() {
		err := k8sHandle.Stop()
		So(err, ShouldBeNil)
		err = k8sHandle.Clean()
		So(err, ShouldBeNil)
		err = k8sHandle.EraseOutput()
		So(err, ShouldBeNil)
	}()

	k8sHandle.Wait(0 * time.Millisecond)
}
