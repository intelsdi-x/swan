package executor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/kubernetes"
	"github.com/intelsdi-x/athena/pkg/utils/fs"
	"github.com/nu7hatch/gouuid"
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

func TestKubernetesExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("Creating a kubernetes executor _with_ a kubernetes cluster available", t, func() {
		local := executor.NewLocal()

		config, err := kubernetes.UniqueConfig()
		So(err, ShouldBeNil)

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
		So(len(strings.Split(out, "\n")), ShouldEqual, 1)

		// Skipping for now as the Kopernik CI breaks here.
		SkipConvey("The generic Executor test should pass", func() {
			testExecutor(t, k8sexecutor)
		})

		Convey("Running a command with a successful exit status should leave one pod running", func() {
			taskHandle, err := k8sexecutor.Execute("sleep 4 && exit 0")
			So(err, ShouldBeNil)
			defer taskHandle.EraseOutput()
			defer taskHandle.Clean()
			defer taskHandle.Stop()

			out, err := kubectl(executorConfig.Address, "get pods")
			So(err, ShouldBeNil)
			So(len(strings.Split(out, "\n")), ShouldEqual, 2)

			Convey("And after at most 5 seconds", func() {
				So(taskHandle.Wait(5*time.Second), ShouldBeTrue)

				Convey("The exit status should be zero", func() {
					exitCode, err := taskHandle.ExitCode()
					So(err, ShouldBeNil)
					So(exitCode, ShouldEqual, 0)

					Convey("And there should be zero pods", func() {
						out, err = kubectl(executorConfig.Address, "get pods")
						So(err, ShouldBeNil)
						So(len(strings.Split(out, "\n")), ShouldEqual, 1)
					})
				})
			})
		})

		Convey("Running a command with an unsuccessful exit status should leave one pod running", func() {
			taskHandle, err := k8sexecutor.Execute("sleep 3 && exit 5")
			So(err, ShouldBeNil)
			defer taskHandle.EraseOutput()
			defer taskHandle.Clean()
			defer taskHandle.Stop()

			out, err := kubectl(executorConfig.Address, "get pods")
			So(err, ShouldBeNil)
			So(len(strings.Split(out, "\n")), ShouldEqual, 2)

			Convey("And after at most 5 seconds", func() {
				So(taskHandle.Wait(5*time.Second), ShouldBeTrue)

				Convey("The exit status should be 5", func() {
					exitCode, err := taskHandle.ExitCode()
					So(err, ShouldBeNil)
					So(exitCode, ShouldEqual, 5)

					Convey("And there should be zero pods", func() {
						out, err = kubectl(executorConfig.Address, "get pods")
						So(err, ShouldBeNil)
						So(len(strings.Split(out, "\n")), ShouldEqual, 1)
					})
				})
			})
		})

		Convey("Running a command and calling Clean() on task handle should not cause a data race", func() {
			taskHandle, err := k8sexecutor.Execute("sleep 3 && exit 0")
			So(err, ShouldBeNil)
			taskHandle.Clean()

			defer taskHandle.EraseOutput()
			defer taskHandle.Stop()
		})

		Convey("Logs should be available and non-empty", func() {
			taskHandle, err := k8sexecutor.Execute("echo \"This is Sparta\" && (echo \"This is England\" 1>&2) && exit 0")
			So(err, ShouldBeNil)
			So(taskHandle.Wait(5*time.Second), ShouldBeTrue)
			time.Sleep(10 * time.Second)

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

			defer taskHandle.EraseOutput()
			defer taskHandle.Clean()
			defer taskHandle.Stop()
		})

		Convey("Timeout should not block execution because of files being unavailable", func() {
			client, err := client.New(&restclient.Config{
				Host:     executorConfig.Address,
				Username: executorConfig.Username,
				Password: executorConfig.Password,
			})
			So(err, ShouldBeNil)
			nodes, err := client.Nodes().List(api.ListOptions{})
			So(err, ShouldBeNil)
			node := nodes.Items[0]
			newTaint := api.Taint{
				Key: "hponly", Value: "true", Effect: api.TaintEffectNoSchedule,
			}
			taintsInJSON, err := json.Marshal([]api.Taint{newTaint})
			So(err, ShouldBeNil)
			node.Annotations[api.TaintsAnnotationKey] = string(taintsInJSON)
			_, err = client.Nodes().Update(&node)
			So(err, ShouldBeNil)

			executorConfig.LaunchTimeout = 1 * time.Second
			k8sexecutor, err = executor.NewKubernetes(executorConfig)
			So(err, ShouldBeNil)
			taskHandle, err := k8sexecutor.Execute("sleep inf")
			So(err, ShouldBeNil)
			defer taskHandle.EraseOutput()
			defer taskHandle.Clean()
			defer taskHandle.Stop()

			stopped := taskHandle.Wait(5 * time.Second)
			So(stopped, ShouldBeTrue)
		})

	})
}

func kubectl(server string, subcommand string) (string, error) {
	kubectlBinPath := path.Join(fs.GetAthenaBinPath(), "kubectl")
	buf := new(bytes.Buffer)
	cmd := exec.Command("sh", "-c", fmt.Sprintf("%s -s %s %s", kubectlBinPath, server, subcommand))
	cmd.Stdout = buf
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}
