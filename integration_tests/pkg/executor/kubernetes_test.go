package executor

import (
	"fmt"
	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/kubernetes"
	"github.com/nu7hatch/gouuid"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestKubernetesExecutor(t *testing.T) {
	Convey("Creating a kubernetes executor _with_ a kubernetes cluster available", t, func() {
		local := executor.NewLocal()

		// NOTE: To reduce the likelihood of port conflict between test kubernetes clusters, we randomly
		// assign a collection of ports to the services. Eventhough previous kubernetes processes
		// have been shut down, ports may be in CLOSE_WAIT state.
		config := kubernetes.DefaultConfig()
		ports := testhelpers.RandomPorts(36000, 40000, 5)
		So(len(ports), ShouldEqual, 5)
		config.KubeAPIPort = ports[0]
		config.KubeletPort = ports[1]
		config.KubeControllerPort = ports[2]
		config.KubeSchedulerPort = ports[3]
		config.KubeProxyPort = ports[4]

		k8sLauncher := kubernetes.New(local, local, config)
		So(k8sLauncher, ShouldNotBeNil)

		k8sHandle, err := k8sLauncher.Launch()
		So(err, ShouldBeNil)

		// Make sure cluster is shut down and cleaned up when test ends.
		defer func() {
			var errors []string
			err := k8sHandle.Stop()
			if err != nil {
				t.Logf(err.Error())
				errors = append(errors, err.Error())
			}

			err = k8sHandle.Clean()
			if err != nil {
				t.Logf(err.Error())
				errors = append(errors, err.Error())
			}

			err = k8sHandle.EraseOutput()
			if err != nil {
				t.Logf(err.Error())
				errors = append(errors, err.Error())
			}

			So(len(errors), ShouldEqual, 0)
		}()

		podName, err := uuid.NewV4()
		So(err, ShouldBeNil)

		executorConfig := executor.DefaultKubernetesConfig()
		executorConfig.Address = fmt.Sprintf("http://127.0.0.1:%d", config.KubeAPIPort)
		executorConfig.PodName = podName.String()
		k8sexecutor, err := executor.NewKubernetes(executorConfig)
		So(err, ShouldBeNil)

		Convey("Running a command with a successful exit status", func() {
			taskHandle, err := k8sexecutor.Execute("sleep 1 && exit 0")
			So(err, ShouldBeNil)

			// TODO: Verify that pod is running using kubectl (we have one second to do this).

			Convey("And after at most 5 seconds", func() {
				So(taskHandle.Wait(5*time.Second), ShouldBeTrue)

				Convey("The exit status should be zero", func() {
					exitCode, err := taskHandle.ExitCode()
					So(err, ShouldBeNil)
					So(exitCode, ShouldEqual, 0)
				})
			})
		})

		Convey("Running a command with an unsuccessful exit status", func() {
			taskHandle, err := k8sexecutor.Execute("sleep 1 && exit 5")
			So(err, ShouldBeNil)

			Convey("And after at most 5 seconds", func() {
				So(taskHandle.Wait(5*time.Second), ShouldBeTrue)

				Convey("The exit status should be 5", func() {
					exitCode, err := taskHandle.ExitCode()
					So(err, ShouldBeNil)
					So(exitCode, ShouldEqual, 5)
				})
			})
		})
	})
}
