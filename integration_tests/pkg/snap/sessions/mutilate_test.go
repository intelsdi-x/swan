package sessions

import (
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	snapTest "github.com/intelsdi-x/swan/integration_tests/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/executor/mocks"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions"
	. "github.com/smartystreets/goconvey/convey"
)

func setupMutilateExpectedMetrics() map[string]string {
	expectedMetrics := make(map[string]string)
	// These are results from test output file
	// in "src/github.com/intelsdi-x/swan/misc/
	// snap-plugin-collector-mutilate/mutilate/mutilate.stdout"
	expectedMetrics["avg"] = "20.80000"
	expectedMetrics["std"] = "23.10000"
	expectedMetrics["min"] = "11.90000"
	expectedMetrics["5th"] = "13.30000"
	expectedMetrics["10th"] = "13.40000"
	expectedMetrics["90th"] = "33.40000"
	expectedMetrics["95th"] = "43.10000"
	expectedMetrics["99th"] = "59.50000"

	return expectedMetrics
}

func validateReturnedMetricRow(expectedMetrics map[string]string, namespace string,
	tags string, value string) {

	// Check tags.
	tagsSplitted := strings.Split(tags, ",")
	So(len(tagsSplitted), ShouldBeGreaterThan, 2)

	// Unfortunately we are not sure about the order in this slice.
	So(tagsSplitted[0], ShouldBeIn,
		"swan_experiment=foobar", "swan_phase=barbaz", "swan_repetition=1")
	So(tagsSplitted[1], ShouldBeIn,
		"swan_experiment=foobar", "swan_phase=barbaz", "swan_repetition=1")
	So(tagsSplitted[2], ShouldBeIn,
		"swan_experiment=foobar", "swan_phase=barbaz", "swan_repetition=1")
	logrus.SetLevel(logrus.ErrorLevel)
	logrus.Error("Checking ", namespace)
	// Check namespace & values.
	namespaceSplitted := strings.Split(namespace, "/")
	expectedValue, ok := expectedMetrics[namespaceSplitted[len(namespaceSplitted)-1]]
	So(ok, ShouldBeTrue)
	So(expectedValue, ShouldEqual, value)
}

func TestMutilateSnapSession(t *testing.T) {
	var snapd *snapTest.Snapd
	var c *client.Client
	var publisher *wmap.PublishWorkflowMapNode
	var metricsFile string
	expectedMetrics := setupMutilateExpectedMetrics()

	goPath := os.Getenv("GOPATH")
	buildPath := path.Join(goPath, "src", "github.com", "intelsdi-x", "swan", "build")

	Convey("While having Snapd running", t, func() {
		snapd = snapTest.NewSnapd()
		snapd.Execute()

		// Wait until snap is up.
		So(snapd.Connected(), ShouldBeTrue)

		defer func() {
			if snapd != nil {
				snapd.Stop()
				snapd.CleanAndEraseOutput()
			}
		}()

		Convey("We are able to connect with snapd", func() {
			ct, err := client.New("http://127.0.0.1:8181", "v1", true)

			Convey("Shouldn't return any errors", func() {
				So(err, ShouldBeNil)
			})

			c = ct

			Convey("Loading test publisher", func() {
				plugins := snap.NewPlugins(c)
				So(plugins, ShouldNotBeNil)

				pluginPath := []string{path.Join(buildPath, "snap-plugin-publisher-session-test")}
				plugins.Load(pluginPath)

				publisher = wmap.NewPublishNode("session-test", 1)

				So(publisher, ShouldNotBeNil)

				tmpFile, err := ioutil.TempFile("", "session_test")
				tmpFile.Close()
				So(err, ShouldBeNil)

				metricsFile = tmpFile.Name()

				publisher.AddConfigItem("file", metricsFile)

				Convey("While launching MutilateSnapSession", func() {
					mutilateSnapSession := sessions.NewMutilateSnapSessionLauncher(
						buildPath, 1*time.Second, c, publisher,
					)

					mockedTaskInfo := new(mocks.TaskInfo)
					mutilateStdoutPath := path.Join(goPath, "src/github.com/intelsdi-x/swan/misc/snap-plugin-collector-mutilate/mutilate/mutilate.stdout")

					file, err := os.Open(mutilateStdoutPath)

					So(err, ShouldBeNil)
					defer file.Close()

					mockedTaskInfo.On("StdoutFile").Return(file, nil)
					customPhaseSession := phase.Session{
						ExperimentID: "foobar",
						PhaseID:      "barbaz",
						RepetitionID: 1,
					}

					handle, err := mutilateSnapSession.Launch(mockedTaskInfo, customPhaseSession)
					So(err, ShouldBeNil)

					defer func() {
						err := handle.Stop()
						So(err, ShouldBeNil)
					}()

					Convey("Contacting snap to get the task status", func() {
						status, err := handle.Status()
						So(err, ShouldBeNil)

						So(status, ShouldEqual, "Running")

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
										// Debug only. Testing bug on bad slice order.
										// SCE-328.
										logrus.Errorf(lines[i])
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
										validateReturnedMetricRow(
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
