// +build sequential

package executor

import (
	"encoding/json"
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/intelsdi-x/athena/integration_tests/test_helpers"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/kubernetes"
	"github.com/nu7hatch/gouuid"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/kubernetes/pkg/api"
)

const (
	clusterSpawnTimeout = 2 * time.Minute
	podSpawnTimeout     = 10 * time.Second
	podFinishedTimeout  = 90 * time.Second
)

func TestKubernetesExecutor(t *testing.T) {
	// Create Kubernetes configuration.
	config := kubernetes.DefaultConfig()
	executorConfig := executor.DefaultKubernetesConfig()

	// Create kubectl helper for communicate with Kubernetes cluster.
	kubectl, err := testhelpers.NewKubeClient(executorConfig)
	if err != nil {
		// NewKubeClient() returns error only when kubernetes configuration is
		// invalid.
		t.Errorf("Requested configuration is invalid: %q", err)
	}

	// Create Kubernetes launcher and spawn Kubernetes cluster.
	local := executor.NewLocal()
	k8sLauncher := kubernetes.New(local, local, config)
	k8sHandle, err := k8sLauncher.Launch()
	if err != nil {
		t.Errorf("Cannot start k8s cluster: %q", err)
	}

	// Wait for at least one node is up and running in cluster.
	if err := kubectl.WaitForCluster(clusterSpawnTimeout); err != nil {
		t.Errorf("Cannot launch K8s cluster: %q", err)
	}

	// Make sure cluster is shut down and cleaned up when test ends.
	defer func() {
		errs := executor.StopCleanAndErase(k8sHandle)

		errs.Add(exec.Command("etcdctl", "rm", "--recursive", "--dir", "/registry").Run())

		if err := errs.GetErrIfAny(); err != nil {
			t.Errorf("Cannot stop cluster: %q", err)
		}
	}()

	Convey("Creating a kubernetes executor _with_ a kubernetes cluster available", t, func() {

		// Generate random pod name. This pod name should be unique for each
		// test case inside this Convey.
		podName, err := uuid.NewV4()
		So(err, ShouldBeNil)
		executorConfig.PodNamePrefix = podName.String()

		// Create Kubernetes executor, which should be passed to following
		// Conveys.
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
			defer executor.StopCleanAndErase(taskHandle)
			So(err, ShouldBeNil)

			// Spawning pods on Kubernetes is a complex process which consists
			// of i.a.: scheduling and pulling image. Test should wait for
			// processing executing request.
			err = kubectl.WaitForPod(podSpawnTimeout)
			So(err, ShouldBeNil)

			Convey("And after few seconds", func() {
				// Pod should end after three seconds, but propagation of
				// status information can take longer time. To reduce number
				// of false-positive assertion fails, Wait() timeout is much
				// longer then time withing pod should shutdown.
				So(taskHandle.Wait(podFinishedTimeout), ShouldBeTrue)

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
			defer executor.StopCleanAndErase(taskHandle)
			So(err, ShouldBeNil)

			err = kubectl.WaitForPod(podSpawnTimeout)
			So(err, ShouldBeNil)

			Convey("And after few seconds", func() {
				So(taskHandle.Wait(podFinishedTimeout), ShouldBeTrue)

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
			defer executor.StopCleanAndErase(taskHandle)
			So(err, ShouldBeNil)
			taskHandle.Clean()
		})

		Convey("Logs should be available and non-empty", func() {
			taskHandle, err := k8sexecutor.Execute("echo \"This is Sparta\" && (echo \"This is England\" 1>&2) && exit 0")
			defer executor.StopCleanAndErase(taskHandle)

			So(err, ShouldBeNil)
			So(taskHandle.Wait(podFinishedTimeout), ShouldBeTrue)

			exitCode, err := taskHandle.ExitCode()
			So(exitCode, ShouldEqual, 0)

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

			stderr, err := taskHandle.StderrFile()
			So(err, ShouldBeNil)
			defer stderr.Close()
			buffer = make([]byte, 10)
			n, err = stderr.Read(buffer)

			// stderr will always be empty as we are not able to fetch it from K8s.
			// stdout includes both stderr and stdout of the application run in the pod.
			So(err, ShouldEqual, io.EOF)
			So(n, ShouldEqual, 0)
		})

		Convey("Timeout should not block execution because of files being unavailable", func() {
			nodes, err := kubectl.Nodes().List(api.ListOptions{})
			So(err, ShouldBeNil)
			node := nodes.Items[0]
			newTaint := api.Taint{
				Key: "hponly", Value: "true", Effect: api.TaintEffectNoSchedule,
			}
			taintsInJSON, err := json.Marshal([]api.Taint{newTaint})
			So(err, ShouldBeNil)
			node.Annotations[api.TaintsAnnotationKey] = string(taintsInJSON)
			_, err = kubectl.Client.Nodes().Update(&node)
			So(err, ShouldBeNil)

			executorConfig.LaunchTimeout = 1 * time.Second
			k8sexecutor, err = executor.NewKubernetes(executorConfig)
			So(err, ShouldBeNil)
			taskHandle, err := k8sexecutor.Execute("sleep inf")
			defer executor.StopCleanAndErase(taskHandle)
			So(err, ShouldBeNil)

			stopped := taskHandle.Wait(5 * time.Second)
			So(stopped, ShouldBeTrue)
		})

	})
}
