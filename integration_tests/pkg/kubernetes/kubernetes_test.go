package kubernetes

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/kubernetes"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	"github.com/nu7hatch/gouuid"
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
			config, err := kubernetes.UniqueConfig()
			So(err, ShouldBeNil)

			kubernetesAddress := fmt.Sprintf("http://127.0.0.1:%d", config.KubeAPIPort)

			k8sLauncher := kubernetes.New(local, local, config)
			So(k8sLauncher, ShouldNotBeNil)

			k8sHandle, err := k8sLauncher.Launch()
			So(err, ShouldBeNil)

			defer stopCleanCheckError(k8sHandle)

			Convey("And kubectl shows that local host is in Ready state", func() {
				So(k8sHandle.Wait(100*time.Millisecond), ShouldBeFalse)

				taskHandle, err := local.Execute(fmt.Sprintf("%s -s %s get nodes", kubectlBinPath, kubernetesAddress))
				So(err, ShouldBeNil)

				defer stopCleanCheckError(taskHandle)

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
				// NOTE: We pick a unique deployment name to reduce the likelihood of interference between
				// test runs.
				deploymentName, err := uuid.NewV4()
				So(err, ShouldBeNil)

				// Command kubectl 'run' creates a Deployment with uuid above on Kubernetes cluster.
				podCreateHandle, err := local.Execute(
					fmt.Sprintf("%s -s %s run %s --image=nginx", kubectlBinPath, kubernetesAddress, deploymentName.String()))
				So(err, ShouldBeNil)

				defer stopCleanCheckError(podCreateHandle)

				podCreateHandle.Wait(0)

				file, err := podCreateHandle.StdoutFile()
				So(err, ShouldBeNil)
				So(file, ShouldNotBeNil)

				data, readErr := ioutil.ReadFile(file.Name())
				So(readErr, ShouldBeNil)

				// Output should equal:
				// deployment "<uuid of deployment>" created
				So(string(data), ShouldEqual, fmt.Sprintf("deployment \"%s\" created\n", deploymentName.String()))

				//Remove created pod.
				podRemoveHandle, err := local.Execute(fmt.Sprintf("%s -s %s delete deployment %s", kubectlBinPath, kubernetesAddress, deploymentName.String()))
				So(err, ShouldBeNil)

				defer stopCleanCheckError(podRemoveHandle)

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
