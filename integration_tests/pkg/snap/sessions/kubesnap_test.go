package sessions

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
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
	"github.com/nu7hatch/gouuid"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSnapKubesnapSession(t *testing.T) {
	var snapd *testhelpers.Snapd

	Convey("While having Snapd running", t, func() {
		snapd = testhelpers.NewSnapd()
		err := snapd.Start()
		So(err, ShouldBeNil)

		defer snapd.CleanAndEraseOutput()
		defer snapd.Stop()

		// Wait until snap is up.
		So(snapd.Connected(), ShouldBeTrue)
		snapdAddress := fmt.Sprintf("http://%s:%d", "127.0.0.1", snapd.Port())

		// Load plugins.
		loaderConfig := snap.DefaultPluginLoaderConfig()
		loaderConfig.SnapdAddress = snapdAddress
		loader, err := snap.NewPluginLoader(loaderConfig)
		So(err, ShouldBeNil)

		err = loader.LoadPlugins(snap.KubesnapDockerCollector, snap.SessionPublisher)
		So(err, ShouldBeNil)
		publisherPluginName, _, err := snap.GetPluginNameAndType(snap.SessionPublisher)

		tmpFile, err := ioutil.TempFile("", "session_test")
		So(err, ShouldBeNil)
		tmpFileName := tmpFile.Name()
		tmpFile.Close()
		defer os.Remove(tmpFileName)

		resultFile := tmpFile.Name()
		publisher := wmap.NewPublishNode(publisherPluginName, snap.PluginAnyVersion)
		So(publisher, ShouldNotBeNil)
		publisher.AddConfigItem("file", resultFile)

		// Run Kubernetes
		exec := executor.NewLocal()
		kubernetesLauncher := kubernetes.New(exec, exec, kubernetes.DefaultConfig())
		kubernetesHandle, err := kubernetesLauncher.Launch()
		So(err, ShouldBeNil)
		defer kubernetesHandle.EraseOutput()
		defer kubernetesHandle.Clean()
		defer kubernetesHandle.Stop()

		// Waiting for Kubernetes Executor.
		kubeExecutor, err := executor.NewKubernetes(executor.DefaultKubernetesConfig())
		So(err, ShouldBeNil)

		podHandle, err := kubeExecutor.Execute("stress -c 1 -t 600")
		So(err, ShouldBeNil)
		defer podHandle.EraseOutput()
		defer podHandle.Clean()
		defer podHandle.Stop()

		// Run Prepare Kubesnap Session.
		experimentID, err := uuid.NewV4()
		So(err, ShouldBeNil)
		phaseID, err := uuid.NewV4()
		So(err, ShouldBeNil)

		kubesnapConfig := kubesnap.DefaultConfig()
		kubesnapConfig.SnapdAddress = snapdAddress
		kubesnapConfig.Publisher = publisher
		kubesnapLauncher, err := kubesnap.NewSessionLauncher(kubesnapConfig)
		So(err, ShouldBeNil)
		kubesnapHandle, err := kubesnapLauncher.LaunchSession(
			nil,
			phase.Session{
				ExperimentID: experimentID.String(),
				PhaseID:      phaseID.String(),
				RepetitionID: 1,
			},
		)
		So(err, ShouldBeNil)
		So(kubesnapHandle.IsRunning(), ShouldBeTrue)
		kubesnapHandle.Wait()
		time.Sleep(120 * time.Second) // One hit does not always yield results.
		kubesnapHandle.Stop()

		// Check results here.
		content, err := ioutil.ReadFile(tmpFileName)
		So(err, ShouldBeNil)
		logrus.Errorf("Content: %q\n", string(content))
		logrus.Errorf("Filename: %q\n", tmpFileName)
		So(string(content), ShouldNotEqual, "")

		// Check CPU total usage for container.
		cpuStatsRegex := regexp.MustCompile(`/intel/docker/\S+/cgroups/cpu_stats/cpu_usage/total_usage\s+\S+\s+(\d+)`)
		cpuStatsMatches := cpuStatsRegex.FindStringSubmatch(string(content))
		logrus.Errorf("cpuMatches: %+v", cpuStatsMatches)
		So(len(cpuStatsMatches), ShouldBeGreaterThanOrEqualTo, 2)
		cpuUsage, err := strconv.Atoi(cpuStatsMatches[1])
		So(err, ShouldBeNil)
		So(cpuUsage, ShouldBeGreaterThan, 0)

		// Check Memory usage for container.
		memoryUsageRegex := regexp.MustCompile(`/intel/docker/\S+/cgroups/memory_stats/usage/usage\s+\S+\s+(\d+)`)
		memoryUsageMatches := memoryUsageRegex.FindStringSubmatch(string(content))
		logrus.Errorf("memoryUsageMatches: %+v", cpuStatsMatches)
		So(len(memoryUsageMatches), ShouldBeGreaterThanOrEqualTo, 2)
		memoryUsage, err := strconv.Atoi(memoryUsageMatches[1])
		So(err, ShouldBeNil)
		So(memoryUsage, ShouldBeGreaterThan, 0)
	})
}
