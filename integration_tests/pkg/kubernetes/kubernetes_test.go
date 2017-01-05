// +build sequential

package kubernetes

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/kubernetes"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	kubectlBinPath = path.Join(fs.GetAthenaBinPath(), "kubectl")
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

			defer executor.StopCleanAndErase(k8sHandle)

			Convey("And kubectl shows that local host is in Ready state", func() {
				So(k8sHandle.Wait(100*time.Millisecond), ShouldBeFalse)

				output, err := exec.Command(kubectlBinPath, "-s", kubernetesAddress, "get", "nodes").Output()
				So(err, ShouldBeNil)

				host, err := os.Hostname()
				So(err, ShouldBeNil)

				// kubectl get nodes should return this:
				// NAME            STATUS    AGE
				// <hostname>      Ready     <x>h

				re, err := regexp.Compile(fmt.Sprintf("%s.*?Ready", host))
				So(err, ShouldBeNil)

				match := re.Find(output)
				So(match, ShouldNotBeNil)

				Convey("And K8s cluster can be stopped with no error", func() {
					var errors []string
					err := k8sHandle.Stop()
					if err != nil {
						errors = append(errors, err.Error())
					}
					err = k8sHandle.Clean()
					if err != nil {
						errors = append(errors, err.Error())
					}
					err = k8sHandle.EraseOutput()
					if err != nil {
						errors = append(errors, err.Error())
					}

					So(len(errors), ShouldEqual, 0)
				})

				Convey("And we are able to create and remove pod", func() {
					const containerName = "kubernetes-test-go-stress"
					config := executor.DefaultKubernetesConfig()
					config.Address = kubernetesAddress
					config.ContainerImage = "jess/stress"
					config.ContainerName = containerName
					k8sExecutor, err := executor.NewKubernetes(config)
					So(err, ShouldBeNil)

					podHandle, err := k8sExecutor.Execute("stress -c 1")
					So(err, ShouldBeNil)
					defer executor.StopCleanAndErase(podHandle)

					output, err := exec.Command("sudo", "docker", "ps").Output()
					So(err, ShouldBeNil)
					So(string(output), ShouldContainSubstring, containerName)

					err = podHandle.Stop()
					So(err, ShouldBeNil)

					output, err = exec.Command("sudo", "docker", "ps").Output()
					So(err, ShouldBeNil)
					So(string(output), ShouldNotContainSubstring, containerName)

				})
			})
		})
	})
}
