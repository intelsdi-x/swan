package executor

import (
	"bytes"
	"fmt"
	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/kubernetes"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	"github.com/nu7hatch/gouuid"
	. "github.com/smartystreets/goconvey/convey"
	"os/exec"
	"path"
	"strings"
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

		// Make sure no pods are running. Output from kubectl includes a header line. Therefore, with
		// no pod entry, we expect a line count of 1.
		out, err := kubectl(executorConfig.Address, "get pods")
		So(err, ShouldBeNil)
		t.Logf(out)
		So(len(strings.Split(out, "\n")), ShouldEqual, 1)

		Convey("Running a command with a successful exit status should leave one pod running", func() {
			taskHandle, err := k8sexecutor.Execute("sleep 2 && exit 0")
			So(err, ShouldBeNil)

			out, err := kubectl(executorConfig.Address, "get pods")
			So(err, ShouldBeNil)
			t.Logf(out)
			So(len(strings.Split(out, "\n")), ShouldEqual, 2)

			Convey("And after at most 5 seconds", func() {
				So(taskHandle.Wait(5*time.Second), ShouldBeTrue)

				Convey("The exit status should be zero", func() {
					exitCode, err := taskHandle.ExitCode()
					So(err, ShouldBeNil)
					So(exitCode, ShouldEqual, 0)
				})

				Convey("And there should be zero pods", func() {
					out, err = kubectl(executorConfig.Address, "get pods")
					t.Logf(out)
					So(err, ShouldBeNil)
					So(len(strings.Split(out, "\n")), ShouldEqual, 1)
				})
			})
		})

		Convey("Running a command with an unsuccessful exit status should leave one pod running", func() {
			taskHandle, err := k8sexecutor.Execute("sleep 2 && exit 5")
			So(err, ShouldBeNil)

			out, err := kubectl(executorConfig.Address, "get pods")
			t.Logf(out)
			So(err, ShouldBeNil)
			So(len(strings.Split(out, "\n")), ShouldEqual, 2)

			Convey("And after at most 5 seconds", func() {
				So(taskHandle.Wait(5*time.Second), ShouldBeTrue)

				Convey("The exit status should be 5", func() {
					exitCode, err := taskHandle.ExitCode()
					So(err, ShouldBeNil)
					So(exitCode, ShouldEqual, 5)
				})

				Convey("And there should be zero pods", func() {
					out, err = kubectl(executorConfig.Address, "get pods")
					t.Logf(out)
					So(err, ShouldBeNil)
					So(len(strings.Split(out, "\n")), ShouldEqual, 1)
				})
			})
		})
	})
}

func kubectl(server string, subcommand string) (string, error) {
	kubectlBinPath := path.Join(fs.GetSwanBinPath(), "kubectl")
	buf := new(bytes.Buffer)
	cmd := exec.Command("sh", "-c", fmt.Sprintf("%s -s %s %s", kubectlBinPath, server, subcommand))
	cmd.Stdout = buf
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}
