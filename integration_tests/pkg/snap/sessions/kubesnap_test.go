package sessions

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/intelsdi-x/athena/integration_tests/test_helpers"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/kubernetes"
	"github.com/intelsdi-x/athena/pkg/snap"
	"github.com/intelsdi-x/athena/pkg/snap/sessions/kubesnap"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSnapKubesnapSession(t *testing.T) {
	Convey("Preparing Snap and Kubernetes enviroment", t, func() {
		snapd := testhelpers.NewSnapd()
		err := snapd.Start()
		So(err, ShouldBeNil)

		defer snapd.CleanAndEraseOutput()
		defer snapd.Stop()

		snapdAddress := fmt.Sprintf("http://%s:%d", "127.0.0.1", snapd.Port())

		// Load plugins.
		loaderConfig := snap.DefaultPluginLoaderConfig()
		loaderConfig.SnapdAddress = snapdAddress
		loader, err := snap.NewPluginLoader(loaderConfig)
		So(err, ShouldBeNil)

		err = loader.Load(snap.KubesnapDockerCollector, snap.FilePublisher)
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
		config, err := kubernetes.UniqueConfig()
		So(err, ShouldBeNil)
		kubernetesLauncher := kubernetes.New(exec, exec, config)
		kubernetesHandle, err := kubernetesLauncher.Launch()
		So(err, ShouldBeNil)
		defer kubernetesHandle.EraseOutput()
		defer kubernetesHandle.Clean()
		defer kubernetesHandle.Stop()

		// Waiting for Kubernetes Executor.
		kubernetesConfig := executor.DefaultKubernetesConfig()
		kubernetesConfig.Address = fmt.Sprintf("127.0.0.1:%d", config.KubeAPIPort)
		kubeExecutor, err := executor.NewKubernetes(kubernetesConfig)
		So(err, ShouldBeNil)

		podHandle, err := kubeExecutor.Execute("stress -c 1 -t 600")
		So(err, ShouldBeNil)
		defer podHandle.EraseOutput()
		defer podHandle.Clean()
		defer podHandle.Stop()

		// Prepare Kubesnap Session.
		/*experimentID, err := uuid.NewV4()
		So(err, ShouldBeNil)
		phaseID, err := uuid.NewV4()
		So(err, ShouldBeNil)*/

		Convey("Launching Kubesnap Session", func() {
			kubesnapConfig := kubesnap.DefaultConfig()
			kubesnapConfig.SnapdAddress = snapdAddress
			kubesnapConfig.Publisher = publisher
			kubesnapLauncher, err := kubesnap.NewSessionLauncher(kubesnapConfig)
			So(err, ShouldBeNil)
			kubesnapHandle, err := kubesnapLauncher.LaunchSession(
				nil,
				/*phase.Session{
					ExperimentID: experimentID.String(),
					PhaseID:      phaseID.String(),
					RepetitionID: 1,
				},*/
				"foo:bar",
			)
			So(err, ShouldBeNil)
			So(kubesnapHandle.IsRunning(), ShouldBeTrue)
			kubesnapHandle.Wait()
			time.Sleep(5 * time.Second) // One hit does not always yield results.
			kubesnapHandle.Stop()

			// Check results here.
			content, err := ioutil.ReadFile(resultsFileName)
			So(err, ShouldBeNil)
			So(string(content), ShouldNotBeEmpty)

			Convey("There should be CPU results of docker containers on Kubernetes", func() {
				cpuStatsRegex := regexp.MustCompile(`/intel/docker/\S+/cgroups/cpu_stats/cpu_usage/total_usage\|(\d+)`)
				cpuStatsMatches := cpuStatsRegex.FindStringSubmatch(string(content))
				So(len(cpuStatsMatches), ShouldBeGreaterThanOrEqualTo, 2)
				cpuUsage, err := strconv.Atoi(cpuStatsMatches[1])
				So(err, ShouldBeNil)
				So(cpuUsage, ShouldBeGreaterThan, 0)

				Convey("There should be Memory results of docker containers on Kubernetes", func() {
					memoryUsageRegex := regexp.MustCompile(`/intel/docker/\S+/cgroups/memory_stats/usage/usage\|(\d+)`)
					memoryUsageMatches := memoryUsageRegex.FindStringSubmatch(string(content))
					So(len(memoryUsageMatches), ShouldBeGreaterThanOrEqualTo, 2)
					memoryUsage, err := strconv.Atoi(memoryUsageMatches[1])
					So(err, ShouldBeNil)
					So(memoryUsage, ShouldBeGreaterThan, 0)
				})
			})
		})
	})
}
