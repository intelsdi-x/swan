package kubernetes

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/kubernetes"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"testing"
	"time"
)

var (
	kubectlBinPath = path.Join(fs.GetSwanBinPath(), "kubectl")
)

// Please see `pkg/kubernetes/README.md` for prerequisites for this test.
func TestLocalKubernetesPodExecution(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)

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

			Convey("And kubectl shows that local host is in Ready state", func() {
				So(k8sHandle.Wait(100*time.Millisecond), ShouldBeFalse)

				taskHandle, err := local.Execute(fmt.Sprintf("%s get nodes", kubectlBinPath))
				So(err, ShouldBeNil)

				defer func() {
					taskHandle.Stop()
					taskHandle.Clean()
					taskHandle.EraseOutput()
				}()

				taskHandle.Wait(0)

				file, err := taskHandle.StdoutFile()
				So(err, ShouldBeNil)
				So(file, ShouldNotBeNil)

				data, readErr := ioutil.ReadAll(file)
				So(readErr, ShouldBeNil)

				host, err := os.Hostname()
				So(err, ShouldBeNil)

				// kubectl get nodes should return this:
				// NAME            STATUS    AGE
				// <hostname>      Ready     <x>h

				re, err := regexp.Compile(fmt.Sprintf("%s.*Ready", host))
				So(err, ShouldBeNil)

				match := re.Find(data)
				So(match, ShouldNotBeNil)
			})

			// TODO(bp): Create pod & remove. Not a part of SCE-504.
		})
	})
}
