// Copyright (c) 2017 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubernetes

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/kubernetes"
	. "github.com/smartystreets/goconvey/convey"
)

// Please see `pkg/kubernetes/README.md` for prerequisites for this test.
func TestLocalKubernetesPodExecution(t *testing.T) {

	hyperkubeBinPath := testhelpers.AssertFileExists("hyperkube")

	logrus.SetLevel(logrus.ErrorLevel)
	Convey("While having local executor", t, func() {

		local := executor.NewLocal()

		Convey("We are able to launch kubernetes cluster on one node", func() {
			config := kubernetes.DefaultConfig()

			kubernetesAddress := fmt.Sprintf("http://127.0.0.1:%d", config.KubeAPIPort)

			k8sLauncher := kubernetes.New(local, local, config)
			So(k8sLauncher, ShouldNotBeNil)

			k8sHandle, err := k8sLauncher.Launch()
			So(err, ShouldBeNil)

			defer executor.StopAndEraseOutput(k8sHandle)

			Convey("And kubectl shows that local host is in Ready state", func() {
				terminated, err := k8sHandle.Wait(100 * time.Millisecond)
				So(err, ShouldBeNil)
				So(terminated, ShouldBeFalse)

				output, err := exec.Command(hyperkubeBinPath, "kubectl", "-s", kubernetesAddress, "get", "nodes").Output()
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
					defer executor.StopAndEraseOutput(podHandle)

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
