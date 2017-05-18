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
	"flag"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/conf"
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
					config.ContainerName = containerName
					k8sExecutor, err := executor.NewKubernetes(config)
					So(err, ShouldBeNil)

					podHandle, err := k8sExecutor.Execute("stress-ng -c 1")
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

// Please see `pkg/kubernetes/README.md` for prerequisites for this test.
func TestLocalKubernetesPodBrokenExecution(t *testing.T) {

	flag.Set("kubernetes_cluster_clean_left_pods_on_startup", "true")

	conf.ParseFlags()

	podName := "swan-testpod"
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

				re, err := regexp.Compile(fmt.Sprintf("%s.*?Ready", host))
				So(err, ShouldBeNil)

				match := re.Find(output)
				So(match, ShouldNotBeNil)

				Convey("And we are able to create pod", func() {
					output, err := exec.Command(hyperkubeBinPath, "kubectl", "-s", kubernetesAddress, "run", podName, "--image=intelsdi/swan", "--restart=Never", "--", "sleep", "inf").Output()
					So(err, ShouldBeNil)

					re, err := regexp.Compile("created")
					So(err, ShouldBeNil)
					match := re.Find(output)
					So(match, ShouldNotBeNil)

					time.Sleep(10 * time.Second)

					Convey("If kubernetes is stopped pod shall be alive", func() {
						_ = executor.StopAndEraseOutput(k8sHandle)

						output, err := exec.Command("sudo", "docker", "ps").Output()
						So(err, ShouldBeNil)
						So(string(output), ShouldContainSubstring, podName)

						Convey("After starting again kubernetes old pod shall be removed", func() {
							k8sHandle, err = k8sLauncher.Launch()
							So(err, ShouldBeNil)

							output, err := exec.Command("sudo", "docker", "ps").Output()
							So(err, ShouldBeNil)
							So(string(output), ShouldNotContainSubstring, podName)

							Convey("kubernetes launcher shall start pod without error", func() {
								config := executor.DefaultKubernetesConfig()
								config.Address = kubernetesAddress
								config.ContainerImage = "intelsdi/swan"
								config.PodName = podName
								k8sExecutor, err := executor.NewKubernetes(config)
								So(err, ShouldBeNil)

								podHandle1, err := k8sExecutor.Execute("sleep inf")
								So(err, ShouldBeNil)
								defer executor.StopAndEraseOutput(podHandle1)
							})
						})
					})
				})
			})
		})
	})
}
