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
			So(k8sLauncher, ShouldNotBeNil)

			k8sHandle, err := k8sLauncher.Launch()
			So(err, ShouldBeNil)

			defer func() {
				stopCleanCheckError(k8sHandle)
			}()

			Convey("And kubectl shows that local host is in Ready state", func() {
				So(k8sHandle.Wait(100*time.Millisecond), ShouldBeFalse)

				taskHandle, err := local.Execute(fmt.Sprintf("%s get nodes", kubectlBinPath))
				So(err, ShouldBeNil)

				defer func() {
					stopCleanCheckError(taskHandle)
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

			Convey("And we are able to create and remove pod", func() {
				// Command kubectl 'run' creates a Deployment named “nginx” on Kubernetes cluster.
				podCreateHandle, err := local.Execute(fmt.Sprintf("%s run test --image=nginx", kubectlBinPath))
				So(err, ShouldBeNil)

				defer func() {
					stopCleanCheckError(podCreateHandle)
				}()

				podCreateHandle.Wait(0)

				file, err := podCreateHandle.StdoutFile()
				So(err, ShouldBeNil)
				So(file, ShouldNotBeNil)

				data, readErr := ioutil.ReadFile(file.Name())
				So(readErr, ShouldBeNil)

				// Output should equal:
				// deployment "test" created
				So(string(data), ShouldEqual, "deployment \"test\" created\n")

				//Remove created pod.
				podRemoveHandle, err := local.Execute(fmt.Sprintf("%s delete deployment test", kubectlBinPath))
				So(err, ShouldBeNil)

				defer func() {
					stopCleanCheckError(podRemoveHandle)
				}()

				podRemoveHandle.Wait(0)
			})
		})
	})
}

func stopCleanCheckError(taskHandle executor.TaskHandle) {
	var errors []string
	err := taskHandle.Stop()
	if err != nil {
		errors = append(errors, err.Error())
	}
	err = taskHandle.Clean()
	if err != nil {
		errors = append(errors, err.Error())
	}
	err = taskHandle.EraseOutput()
	if err != nil {
		errors = append(errors, err.Error())
	}
	So(len(errors), ShouldEqual, 0)
}
