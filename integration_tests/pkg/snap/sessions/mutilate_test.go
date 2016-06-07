package sessions

import (
	"fmt"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/executor/mocks"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions"
	"github.com/intelsdi-x/swan/pkg/swan"
	. "github.com/smartystreets/goconvey/convey"
)

func soMetricRowIsValid(expectedMetrics map[string]string, namespace string,
	tags string, value string) {

	// Check tags.
	tagsSplitted := strings.Split(tags, ",")
	So(len(tagsSplitted), ShouldBeGreaterThan, 2)
	So("swan_experiment=foobar", ShouldBeIn, tagsSplitted)
	So("swan_phase=barbaz", ShouldBeIn, tagsSplitted)
	So("swan_repetition=1", ShouldBeIn, tagsSplitted)

	// Check namespace & values.
	namespaceSplitted := strings.Split(namespace, "/")
	expectedValue, ok := expectedMetrics[namespaceSplitted[len(namespaceSplitted)-1]]
	So(ok, ShouldBeTrue)
	So(expectedValue, ShouldEqual, value)
}

const (
	snapMutilateSessionTestAPIPort = 12346
)

func TestSnapMutilateSession(t *testing.T) {
	var snapd *testhelpers.Snapd
	var publisher *wmap.PublishWorkflowMapNode
	var metricsFile string

	Convey("While having Snapd running", t, func() {
		snapd = testhelpers.NewSnapdOnPort(snapMutilateSessionTestAPIPort)
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

		Convey("We are able to connect with snapd", func() {
			c, err := client.New(
				fmt.Sprintf("http://127.0.0.1:%d", snapMutilateSessionTestAPIPort), "v1", true)
			So(err, ShouldBeNil)

			Convey("Loading test publisher", func() {
				plugins := snap.NewPlugins(c)
				So(plugins, ShouldNotBeNil)

				pluginPath := []string{
					path.Join(swan.GetSwanBuildPath(), "snap-plugin-publisher-session-test"),
				}
				plugins.Load(pluginPath)

				publisher = wmap.NewPublishNode("session-test", 1)

				So(publisher, ShouldNotBeNil)

				tmpFile, err := ioutil.TempFile("", "session_test")
				So(err, ShouldBeNil)
				tmpFile.Close()

				metricsFile = tmpFile.Name()

				publisher.AddConfigItem("file", metricsFile)

				Convey("While launching MutilateSnapSession", func() {
					mutilateSnapSession := sessions.NewMutilateSnapSessionLauncher(
						swan.GetSwanBuildPath(), 1*time.Second, c, publisher,
					)

					mockedTaskInfo := new(mocks.TaskInfo)
					mutilateStdoutPath := path.Join(
						os.Getenv("GOPATH"), "src/github.com/intelsdi-x/swan/misc/snap-plugin-collector-mutilate/mutilate/mutilate.stdout")

					file, err := os.Open(mutilateStdoutPath)

					So(err, ShouldBeNil)
					defer file.Close()

					mockedTaskInfo.On("StdoutFile").Return(file, nil)
					session := phase.Session{
						ExperimentID: "foobar",
						PhaseID:      "barbaz",
						RepetitionID: 1,
					}

					handle, err := mutilateSnapSession.LaunchSession(mockedTaskInfo, session)
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
							"avg":  "20.80000",
							"std":  "23.10000",
							"min":  "11.90000",
							"5th":  "13.30000",
							"10th": "13.40000",
							"90th": "33.40000",
							"95th": "43.10000",
							"99th": "59.50000",
							"qps":  "4993.10000",
						}

						Convey("Reading samples from file", func() {
							retries := 5
							found := false
							for i := 0; i < retries; i++ {
								time.Sleep(500 * time.Millisecond)

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
										soMetricRowIsValid(
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
