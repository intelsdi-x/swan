package snap

import (
	"fmt"

	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
)

const (
	snapSessionTestAPIPort = 12345
)

func TestSnap(t *testing.T) {
	var snapd *testhelpers.Snapd
	var s *snap.Session
	var publisher *wmap.PublishWorkflowMapNode
	var metricsFile string

	goPath := os.Getenv("GOPATH")
	buildPath := path.Join(goPath, "src", "github.com", "intelsdi-x", "swan", "build")

	Convey("While having Snapd running", t, func() {
		snapd = testhelpers.NewSnapdOnPort(snapSessionTestAPIPort)
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
				fmt.Sprintf("http://127.0.0.1:%d", snapSessionTestAPIPort), "v1", true)
			So(err, ShouldBeNil)

			Convey("Loading collectors", func() {
				plugins := snap.NewPlugins(c)
				So(plugins, ShouldNotBeNil)

				pluginPath := []string{path.Join(buildPath, "snap-plugin-collector-session-test")}
				err := plugins.Load(pluginPath)
				So(err, ShouldBeNil)

				// Wait until metric is available in namespace.
				retries := 10
				found := false
				for i := 0; i < retries && !found; i++ {
					m := c.GetMetricCatalog()
					So(m.Err, ShouldBeNil)
					for _, metric := range m.Catalog {
						if metric.Namespace == "/intel/swan/session/metric1" {
							found = true
							break
						}
					}
					time.Sleep(500 * time.Millisecond)
				}
				So(found, ShouldBeTrue)

				Convey("Loading publishers", func() {
					plugins := snap.NewPlugins(c)
					So(plugins, ShouldNotBeNil)

					pluginPath := []string{path.Join(
						buildPath, "snap-plugin-publisher-session-test")}
					err := plugins.Load(pluginPath)
					So(err, ShouldBeNil)

					publisher = wmap.NewPublishNode("session-test", 1)

					So(publisher, ShouldNotBeNil)

					tmpFile, err := ioutil.TempFile("", "session_test")
					tmpFile.Close()
					So(err, ShouldBeNil)

					metricsFile = tmpFile.Name()

					publisher.AddConfigItem("file", metricsFile)

					Convey("While starting a Snap experiment session", func() {
						s = snap.NewSession([]string{"/intel/swan/session/metric1"}, 1*time.Second, c, publisher)
						So(s, ShouldNotBeNil)

						err := s.Start(phase.Session{
							ExperimentID: "foobar",
							PhaseID:      "barbaz",
							RepetitionID: 1,
						})

						So(err, ShouldBeNil)

						defer func() {
							err := s.Stop()
							So(err, ShouldBeNil)
						}()

						Convey("Contacting snap to get the task status", func() {
							status, err := s.Status()
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
										if len(lines) < 1 {
											t.Log("There should be at least one line. Checking again.")
											continue
										}

										columns := strings.Split(lines[0], "\t")
										if len(columns) < 2 {
											t.Log("There should be at least 2 columns. Checking again.")
											continue
										}

										tags := strings.Split(columns[1], ",")
										if len(tags) < 3 {
											t.Log("There should be at least 3 tags. Checking again.")
											continue
										}

										So(columns[0], ShouldEqual, "/intel/swan/session/metric1")
										So("swan_experiment=foobar", ShouldBeIn, tags)
										So("swan_phase=barbaz", ShouldBeIn, tags)
										So("swan_repetition=1", ShouldBeIn, tags)

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
	})
}
