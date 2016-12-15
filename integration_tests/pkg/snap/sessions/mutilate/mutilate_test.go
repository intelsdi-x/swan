package mutilatesessiontest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/intelsdi-x/athena/integration_tests/test_helpers"
	"github.com/intelsdi-x/athena/pkg/executor/mocks"
	"github.com/intelsdi-x/athena/pkg/snap"
	"github.com/intelsdi-x/athena/pkg/snap/sessions/mutilate"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/intelsdi-x/swan/integration_tests/pkg/snap/sessions/helpers"
)

func TestSnapMutilateSession(t *testing.T) {
	var snapd *testhelpers.Snapd
	var publisher *wmap.PublishWorkflowMapNode
	var metricsFile string

	Convey("While having Snapd running", t, func() {
		snapd = testhelpers.NewSnapd()
		err := snapd.Start()
		So(err, ShouldBeNil)

		defer func() {
			if snapd != nil {
				err := snapd.Stop()
				err2 := snapd.CleanAndEraseOutput()

				So(err, ShouldBeNil)
				So(err2, ShouldBeNil)
			}
		}()

		// Wait until snap is up.
		So(snapd.Connected(), ShouldBeTrue)

		snapdAddress := fmt.Sprintf("http://127.0.0.1:%d", snapd.Port())

		loaderConfig := snap.DefaultPluginLoaderConfig()
		loaderConfig.SnapdAddress = snapdAddress
		loader, err := snap.NewPluginLoader(loaderConfig)
		So(err, ShouldBeNil)

		Convey("We are able to connect with snapd", func() {
			Convey("Loading test publisher", func() {
				tmpFile, err := ioutil.TempFile("", "session_test")
				So(err, ShouldBeNil)
				tmpFile.Close()

				metricsFile = tmpFile.Name()
				defer os.Remove(metricsFile)

				loader.Load(snap.SessionPublisher)
				pluginName, _, err := snap.GetPluginNameAndType(snap.SessionPublisher)
				So(err, ShouldBeNil)

				publisher = wmap.NewPublishNode(pluginName, snap.PluginAnyVersion)
				So(publisher, ShouldNotBeNil)

				publisher.AddConfigItem("file", metricsFile)

				Convey("While launching MutilateSnapSession", func() {
					mutilateSessionConfig := mutilatesession.DefaultConfig()
					mutilateSessionConfig.SnapdAddress = snapdAddress
					mutilateSessionConfig.Publisher = publisher
					mutilateSnapSession, err := mutilatesession.NewSessionLauncher(mutilateSessionConfig)
					So(err, ShouldBeNil)

					mockedTaskInfo := new(mocks.TaskInfo)
					mutilateStdoutPath := path.Join(
						os.Getenv("GOPATH"), "src/github.com/intelsdi-x/swan/misc/snap-plugin-collector-mutilate/mutilate/mutilate.stdout")

					file, err := os.Open(mutilateStdoutPath)

					So(err, ShouldBeNil)
					defer file.Close()

					mockedTaskInfo.On("StdoutFile").Return(file, nil)
					/*session := phase.Session{
						ExperimentID: "foobar",
						PhaseID:      "barbaz",
						RepetitionID: 1,
					}*/

					handle, err := mutilateSnapSession.LaunchSession(mockedTaskInfo, "foo:bar")
					So(err, ShouldBeNil)

					defer func() {
						err := handle.Stop()
						So(err, ShouldBeNil)
					}()

					Convey("Contacting snap to get the task status", func() {
						So(handle.IsRunning(), ShouldBeTrue)

						// These are results from test output file
						// in "src/github.com/intelsdi-x/swan/misc/
						// snap-plugin-collector-mutilate/mutilate/mutilate.stdout"
						expectedMetrics := map[string]string{
							"avg":    "20.80000",
							"std":    "23.10000",
							"min":    "11.90000",
							"5th":    "13.30000",
							"10th":   "13.40000",
							"90th":   "33.40000",
							"95th":   "43.10000",
							"99th":   "59.50000",
							"qps":    "4993.10000",
							"custom": "1777.88781",
						}

						Convey("Reading samples from file", func() {
							retries := 50
							found := false
							for i := 0; i < retries; i++ {
								time.Sleep(100 * time.Millisecond)

								dat, err := ioutil.ReadFile(metricsFile)
								if err != nil {
									continue
								}

								if len(dat) > 0 {
									// Look for tag on metric line.
									lines := strings.Split(string(dat), "\n")
									if len(lines) < len(expectedMetrics) {
										t.Log("There should be at least ",
											len(expectedMetrics),
											" lines. Checking again.")
										continue
									}

									allLinesHaveAllColumns := true
									// All lines should have 3 columns.
									for i := 0; i < len(expectedMetrics); i++ {
										columns := strings.Split(lines[i], "\t")
										if len(columns) < 3 {
											allLinesHaveAllColumns = false
										}
									}

									if !allLinesHaveAllColumns {
										t.Log("There should be at least 3 columns for all lines. ",
											"Checking again.")
										continue
									}

									for i := 0; i < len(expectedMetrics); i++ {
										columns := strings.Split(lines[i], "\t")
										helpers.SoMetricRowIsValid(
											expectedMetrics,
											columns[0], columns[1], columns[2])
									}

									found = true

									break
								}

							}
							So(found, ShouldBeTrue)
						})
					})
				})
			})
		})
	})
}
