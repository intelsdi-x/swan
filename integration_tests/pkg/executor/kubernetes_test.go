package executor

import (
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
		k8sLauncher := kubernetes.New(local, local, kubernetes.DefaultConfig())
		So(k8sLauncher, ShouldNotBeNil)

		k8sHandle, err := k8sLauncher.Launch()
		So(err, ShouldBeNil)

		defer func() {
			stopCleanCheckError(k8sHandle)
		}()

		podName, err := uuid.NewV4()
		So(err, ShouldBeNil)

		k8sexecutor, err := executor.NewKubernetesExecutor(
			executor.KubernetesConfig{
				Address: "http://127.0.0.1:8080",
				PodName: podName.String(),
			},
		)

		Convey("Should not fail", func() {
			So(err, ShouldBeNil)

			Convey("Running a command with a successful exit status", func() {
				taskHandle, err := k8sexecutor.Execute("sleep 1 && exit 0")
				So(err, ShouldBeNil)

				Convey("And after at most 2 seconds", func() {
					So(taskHandle.Wait(2*time.Second), ShouldBeTrue)

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

				Convey("And after at most 2 seconds", func() {
					So(taskHandle.Wait(2*time.Second), ShouldBeTrue)

					Convey("The exit status should be 5", func() {
						exitCode, err := taskHandle.ExitCode()
						So(err, ShouldBeNil)
						So(exitCode, ShouldEqual, 5)
					})
				})
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
