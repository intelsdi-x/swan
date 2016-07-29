package sessions

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
	"github.com/intelsdi-x/swan/pkg/kubernetes"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/kubesnap"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSnapKubesnapSession(t *testing.T) {
	var snapd *testhelpers.Snapd
	//var publisher *wmap.PublishWorkflowMapNode
	//var metricsFile string

	Convey("While having Snapd running", t, func() {
		snapd = testhelpers.NewSnapd()
		err := snapd.Start()
		So(err, ShouldBeNil)

		defer snapd.Stop()
		defer snapd.CleanAndEraseOutput()

		// Wait until snap is up.
		So(snapd.Connected(), ShouldBeTrue)
		snapdAddress := fmt.Sprintf("http://%s:%d", "127.0.0.1", snapd.Port())

		// Load plugins.
		loaderConfig := snap.DefaultPluginLoaderConfig()
		loaderConfig.SnapdAddress = snapdAddress
		loader, err := snap.NewPluginLoader(loaderConfig)
		So(err, ShouldBeNil)

		err = loader.Load(snap.KubesnapDockerCollector)
		So(err, ShouldBeNil)

		err = loader.Load(snap.SessionPublisher)
		So(err, ShouldBeNil)
		publisherPluginName, _, err := snap.GetPluginNameAndType(snap.SessionPublisher)

		tmpFile, err := ioutil.TempFile("", "session_test")
		So(err, ShouldBeNil)
		tmpFileName := tmpFile.Name()
		logrus.Errorf("Result file: %q", tmpFileName)
		tmpFile.Close()

		resultFile := tmpFile.Name()
		publisher := wmap.NewPublishNode(publisherPluginName, snap.PluginAnyVersion)
		So(publisher, ShouldNotBeNil)
		publisher.AddConfigItem("file", resultFile)

		// Run Kubernetes
		exec := executor.NewLocal()
		kubernetesLauncher := kubernetes.New(exec, exec, kubernetes.DefaultConfig())
		kubernetesHandle, err := kubernetesLauncher.Launch()
		So(err, ShouldBeNil)
		defer kubernetesHandle.Stop()
		defer kubernetesHandle.Clean()
		defer kubernetesHandle.EraseOutput()

		// Waiting for Kubernetes Executor.
		kubeExecutor, err := executor.NewKubernetes(executor.DefaultKubernetesConfig())
		So(err, ShouldBeNil)

		podHandle, err := kubeExecutor.Execute("1")
		So(err, ShouldNotBeNil)
		defer podHandle.Stop()
		//defer podHandle.Clean() // Panic!
		//defer podHandle.EraseOutput()

		// Run Prepare Kubesnap Session.
		kubesnapConfig := kubesnap.DefaultConfig()
		kubesnapConfig.SnapdAddress = snapdAddress
		kubesnapLauncher, err := kubesnap.NewSessionLauncher(kubesnapConfig)
		So(err, ShouldBeNil)
		kubesnapHandle, err := kubesnapLauncher.LaunchSession(
			nil,
			phase.Session{
				ExperimentID: "foobar",
				PhaseID:      "barbaz",
				RepetitionID: 1,
			},
		)
		So(err, ShouldBeNil)
		So(kubesnapHandle.IsRunning(), ShouldBeTrue)
		//kubesnapHandle.Wait()
		time.Sleep(20 * time.Second)

		kubesnapHandle.Stop()

		// Check results here.

		content, err := ioutil.ReadFile(tmpFileName)
		So(err, ShouldBeNil)
		logrus.Errorf("File content: %q", string(content))
	})
}
