package sessions

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/kubernetes"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/docker"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSnapDockerSession(t *testing.T) {
	Convey("Preparing Snap and Kubernetes enviroment", t, func() {

		cleanup, loader, snapteldAddr := testhelpers.RunAndTestSnaptel()
		defer cleanup()

		err := loader.Load(snap.DockerCollector, snap.FilePublisher)
		So(err, ShouldBeNil)
		publisherPluginName, _, err := snap.GetPluginNameAndType(snap.FilePublisher)
		So(err, ShouldBeNil)

		resultsFile, err := ioutil.TempFile("", "session_test")
		So(err, ShouldBeNil)
		resultsFileName := resultsFile.Name()
		defer os.Remove(resultsFileName)
		resultsFile.Close()

		publisher := wmap.NewPublishNode(publisherPluginName, snap.PluginAnyVersion)
		publisher.AddConfigItem("file", resultsFileName)

		// Run Kubernetes
		exec := executor.NewLocal()
		config := kubernetes.UniqueConfig()
		kubernetesLauncher := kubernetes.New(exec, exec, config)
		kubernetesHandle, err := kubernetesLauncher.Launch()
		So(err, ShouldBeNil)
		So(kubernetesHandle, ShouldNotBeNil)
		defer kubernetesHandle.EraseOutput()
		defer kubernetesHandle.Stop()

		// Waiting for Kubernetes Executor.
		kubernetesConfig := executor.DefaultKubernetesConfig()
		kubernetesConfig.Address = fmt.Sprintf("127.0.0.1:%d", config.KubeAPIPort)
		kubeExecutor, err := executor.NewKubernetes(kubernetesConfig)
		So(err, ShouldBeNil)

		podHandle, err := kubeExecutor.Execute("stress-ng -c 1")
		So(err, ShouldBeNil)
		defer podHandle.EraseOutput()
		defer podHandle.Stop()

		Convey("Launching Docker Session", func() {
			dockerConfig := docker.DefaultConfig()
			dockerConfig.SnapteldAddress = snapteldAddr
			dockerConfig.Publisher = publisher
			dockerLauncher, err := docker.NewSessionLauncher(dockerConfig)
			So(err, ShouldBeNil)
			dockerHandle, err := dockerLauncher.LaunchSession(
				nil,
				"foo:bar",
			)
			So(err, ShouldBeNil)
			So(dockerHandle.IsRunning(), ShouldBeTrue)
			dockerHandle.Wait()
			time.Sleep(5 * time.Second) // One hit does not always yield results.
			dockerHandle.Stop()

			// one measurement should contains more then one metric.
			oneMeasurement, err := testhelpers.GetOneMeasurementFromFile(resultsFileName)
			So(err, ShouldBeNil)
			So(len(oneMeasurement), ShouldBeGreaterThan, 0)
		})
	})
}
