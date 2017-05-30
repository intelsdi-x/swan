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

package executor

import (
	"fmt"
	"strings"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"

	testhelpers "github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/kubernetes"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	podFinishedTimeout = 90 * time.Second
)

func TestKubernetesExecutor(t *testing.T) {
	// Readable, simple, easy to debug, reproducible and reliable testing environment.
	log.SetLevel(log.ErrorLevel)
	// Using timestamps with nanoseconds makes debugging easier.
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, DisableTimestamp: false, TimestampFormat: time.RFC3339Nano})

	config := kubernetes.DefaultConfig()
	config.RetryCount = 10

	// Pod executor config.
	executorConfig := executor.DefaultKubernetesConfig()
	executorConfig.Address = fmt.Sprintf("http://127.0.0.1:%d", config.KubeAPIPort)

	// Create kubectl helper for communicate with Kubernetes cluster.
	kubectl, err := testhelpers.NewKubeClient(executorConfig)
	if err != nil {
		t.Fatalf("Requested configuration is invalid: %q", err)
	}

	// Create Kubernetes launcher and spawn Kubernetes cluster.
	local := executor.NewLocal()
	k8sLauncher := kubernetes.New(local, local, config)
	k8sHandle, err := k8sLauncher.Launch()
	if err != nil {
		t.Fatalf("Cannot start k8s cluster: %q", err)
	}

	// Make sure cluster is shut down and cleaned up when test ends.
	defer func() {
		errs := executor.StopAndEraseOutput(k8sHandle)
		if err := errs.GetErrIfAny(); err != nil {
			t.Fatalf("Cannot stop cluster: %q", err)
		}
	}()

	Convey("Creating a kubernetes executor _with_ a kubernetes cluster available", t, func() {
		// Create Kubernetes executor, which should be passed to following conveys.
		k8sexecutor, err := executor.NewKubernetes(executorConfig)
		So(err, ShouldBeNil)

		// Make sure no pods are running. GetPods() returns running pods and
		// finished pods. We are expecting that there is no running pods on
		// cluster.
		pods, _, err := kubectl.GetPods()
		So(err, ShouldBeNil)
		So(len(pods), ShouldEqual, 0)

		Convey("The generic Executor test should pass", func() {
			testExecutor(t, k8sexecutor)
		})

		Convey("Running a command with a successful exit status should leave one pod running", func() {
			// Start Kubernetes pod which should die after 3 seconds. ExitCode
			// should pass to taskHandle object.
			taskHandle, err := k8sexecutor.Execute("sleep 3 && exit 0")
			So(err, ShouldBeNil)

			defer executor.StopAndEraseOutput(taskHandle)

			Convey("And after few seconds waiting Wait() returns true", func() {
				// Pod should end after three seconds, but propagation of
				// status information can take longer time. To reduce number
				// of false-positive assertion fails, Wait() timeout is much
				// longer then time that pod needs to shutdown.
				terminated, err := taskHandle.Wait(podFinishedTimeout)
				So(err, ShouldBeNil)
				So(terminated, ShouldBeTrue)

				Convey("The exit status should be zero", func() {
					// ExitCode should appears in TaskHandle object after pod
					// termination.
					exitCode, err := taskHandle.ExitCode()
					So(err, ShouldBeNil)
					So(exitCode, ShouldEqual, 0)

					Convey("And there should be zero pods", func() {
						// There shouldn't be any running pods after test
						// executing.
						pods, _, err = kubectl.GetPods()
						So(err, ShouldBeNil)
						So(len(pods), ShouldEqual, 0)
					})
				})
			})
		})

		Convey("Running a command with an unsuccessful exit status should leave one pod running", func() {
			taskHandle, err := k8sexecutor.Execute("sleep 3 && exit 5")
			So(err, ShouldBeNil)
			defer executor.StopAndEraseOutput(taskHandle)

			Convey("And after few seconds", func() {
				terminated, err := taskHandle.Wait(podFinishedTimeout)
				So(err, ShouldBeNil)
				So(terminated, ShouldBeTrue)

				Convey("The exit status should be 5", func() {
					exitCode, err := taskHandle.ExitCode()
					So(err, ShouldBeNil)
					So(exitCode, ShouldEqual, 5)

					Convey("And there should be zero pods", func() {
						pods, _, err = kubectl.GetPods()
						So(err, ShouldBeNil)
						So(len(pods), ShouldEqual, 0)
					})
				})
			})
		})

		Convey("Running a command and calling Clean() on task handle should not cause a data race", func() {
			taskHandle, err := k8sexecutor.Execute("sleep 3 && exit 0")
			So(err, ShouldBeNil)
			defer executor.StopAndEraseOutput(taskHandle)
		})

		Convey("Logs should be available and non-empty", func() {
			taskHandle, err := k8sexecutor.Execute("echo \"This is Sparta\" && (echo \"This is England\" 1>&2) && exit 0")
			defer executor.StopAndEraseOutput(taskHandle)

			So(err, ShouldBeNil)
			terminated, err := taskHandle.Wait(podFinishedTimeout)
			So(err, ShouldBeNil)
			So(terminated, ShouldBeTrue)

			exitCode, err := taskHandle.ExitCode()
			So(exitCode, ShouldEqual, 0)

			// Stdout
			stdout, err := taskHandle.StdoutFile()
			So(err, ShouldBeNil)
			defer stdout.Close()
			buffer := make([]byte, 31)
			n, err := stdout.Read(buffer)

			So(err, ShouldBeNil)
			So(n, ShouldEqual, 31)
			output := strings.Split(string(buffer), "\n")
			So(output, ShouldHaveLength, 3)
			So(output, ShouldContain, "This is Sparta")
			So(output, ShouldContain, "This is England")

			// Stderr
			stderr, err := taskHandle.StderrFile()
			So(err, ShouldBeNil)
			defer stderr.Close()
			buffer = make([]byte, 31)
			n, err = stderr.Read(buffer)

			So(err, ShouldBeNil)
			So(n, ShouldEqual, 31)
			output = strings.Split(string(buffer), "\n")
			So(output, ShouldHaveLength, 3)
			So(output, ShouldContain, "This is Sparta")
			So(output, ShouldContain, "This is England")
		})

		Convey("Long running pod is not deadlocked when deleted externally", func() {
			executorConfig.PodName = "mypod"
			k8sexecutor, err := executor.NewKubernetes(executorConfig)
			So(err, ShouldBeNil)
			th, err := k8sexecutor.Execute("sleep inf")
			So(err, ShouldBeNil)
			defer executor.StopAndEraseOutput(th)
			// Externally delete the pod.
			err = kubectl.DeletePod(executorConfig.PodName)
			So(err, ShouldBeNil)

			// Wait...
			terminated, err := th.Wait(podFinishedTimeout)
			So(err, ShouldBeNil)
			So(terminated, ShouldBeTrue)

			// Exit code expected about killed.
			exitCode, err := th.ExitCode()
			So(err, ShouldBeNil)
			So(exitCode, ShouldEqual, 137)
		})

		Convey("Timeout occurs when image is not found", func() {
			// Launch timeout is needed because pod is left in Pending/Waiting/ImagePullBackOff state.
			executorConfig.ContainerImage = "notexistingone"
			executorConfig.LaunchTimeout = 1 * time.Second
			k8sexecutor, err := executor.NewKubernetes(executorConfig)
			So(err, ShouldBeNil)
			handle, err := k8sexecutor.Execute("wrong command")
			So(err, ShouldNotBeNil)
			So(handle, ShouldBeNil)
		})

		Convey("Timeout should not block execution because of files being unavailable", func() {
			kubectl.TaintNode()
			defer kubectl.UntaintNode()

			executorConfig.LaunchTimeout = 1 * time.Second
			k8sexecutor, err = executor.NewKubernetes(executorConfig)
			So(err, ShouldBeNil)
			taskHandle, err := k8sexecutor.Execute("sleep inf")
			So(err, ShouldNotBeNil)
			So(taskHandle, ShouldBeNil)
		})
	})
}
