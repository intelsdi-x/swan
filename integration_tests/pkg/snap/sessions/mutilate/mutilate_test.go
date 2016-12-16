package mutilatesession_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/intelsdi-x/athena/integration_tests/test_helpers"
	"github.com/intelsdi-x/athena/pkg/executor/mocks"
	"github.com/intelsdi-x/athena/pkg/snap"
	"github.com/intelsdi-x/athena/pkg/snap/sessions/mutilate"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	. "github.com/smartystreets/goconvey/convey"
)

func soMetricRowIsValid(
	expectedMetrics map[string]string,
	namespace, tags, value string) {

	// Check tags.
	tagsSplitted := strings.Split(tags, ",")
	So(len(tagsSplitted), ShouldBeGreaterThanOrEqualTo, 1)
	So("foo=bar", ShouldBeIn, tagsSplitted)

	// Check namespace & values.
	namespaceSplitted := strings.Split(namespace, "/")
	expectedValue, ok := expectedMetrics[namespaceSplitted[len(namespaceSplitted)-1]]
	So(ok, ShouldBeTrue)

	// Reduce string-encoded-float to common precision for comparison.
	expectedValueFloat, err := strconv.ParseFloat(expectedValue, 64)
	So(err, ShouldBeNil)
	valueFloat, err := strconv.ParseFloat(value, 64)
	So(err, ShouldBeNil)

	epsilon := 0.00001
	So(valueFloat, ShouldAlmostEqual, expectedValueFloat, epsilon)
}

func TestSnapMutilateSession(t *testing.T) {
	var snapteld *testhelpers.Snapteld
	var publisher *wmap.PublishWorkflowMapNode
	var metricsFile string

	Convey("While having Snapteld running", t, func() {
		snapteld = testhelpers.NewSnapteld()
		err := snapteld.Start()
		So(err, ShouldBeNil)

		defer func() {
			if snapteld != nil {
				err := snapteld.Stop()
				err2 := snapteld.CleanAndEraseOutput()

				So(err, ShouldBeNil)
				So(err2, ShouldBeNil)
			}
		}()

		// Wait until snap is up.
		So(snapteld.Connected(), ShouldBeTrue)

		snapteldAddress := fmt.Sprintf("http://127.0.0.1:%d", snapteld.Port())

		loaderConfig := snap.DefaultPluginLoaderConfig()
		loaderConfig.SnapteldAddress = snapteldAddress
		loader, err := snap.NewPluginLoader(loaderConfig)
		So(err, ShouldBeNil)

		Convey("We are able to connect with snapteld", func() {
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
					mutilateSessionConfig.SnapteldAddress = snapteldAddress
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
