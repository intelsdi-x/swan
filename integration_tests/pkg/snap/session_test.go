package snap

import (
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/intelsdi-x/swan/pkg/experiment/phase"
	"github.com/intelsdi-x/swan/pkg/snap"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSnap(t *testing.T) {
	var snapd *Snapd
	var c *client.Client
	var s *snap.Session
	var publisher *wmap.PublishWorkflowMapNode
	var metricsFile string

	goPath := os.Getenv("GOPATH")
	buildPath := path.Join(goPath, "src", "github.com", "intelsdi-x", "swan", "build")

	Convey("Testing snap session", t, func() {
		Convey("Starting snapd", func() {
			snapd = NewSnapd()
			snapd.Execute()

			// Wait until snap is up.
			So(snapd.Connected(), ShouldBeTrue)
		})

		Convey("Connecting to snapd", func() {
			ct, err := client.New("http://127.0.0.1:8181", "v1", true)

			Convey("Shouldn't return any errors", func() {
				So(err, ShouldBeNil)
			})

			c = ct
		})

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
		})

		Convey("Loading publishers", func() {
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
		})

		Convey("Creating a Snap experiment session", func() {
			s = snap.NewSession([]string{"/intel/swan/session/metric1"}, 1*time.Second, c, publisher)
			So(s, ShouldNotBeNil)
		})

		Convey("Starting a session", func() {
			So(s, ShouldNotBeNil)
			err := s.Start(phase.Session{
				ExperimentID: "foobar",
				PhaseID:      "barbaz",
				RepetitionID: 1,
			})

			Convey("Shouldn't return any errors", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("Contacting snap to get the task status", func() {
			status, err := s.Status()

			So(err, ShouldBeNil)

			Convey("And the task should be running", func() {
				So(status, ShouldEqual, "Running")
			})
		})

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
					// Unfortunately we are not sure about the order in this slice.
					So(tags[0], ShouldBeIn,
						"swan_experiment=foobar", "swan_phase=barbaz", "swan_repetition=1")
					So(tags[1], ShouldBeIn,
						"swan_experiment=foobar", "swan_phase=barbaz", "swan_repetition=1")
					So(tags[2], ShouldBeIn,
						"swan_experiment=foobar", "swan_phase=barbaz", "swan_repetition=1")

					found = true

					break
				}

			}
			So(found, ShouldBeTrue)
		})

		Convey("Stopping a session", func() {
			So(s, ShouldNotBeNil)
			err := s.Stop()

			So(err, ShouldBeNil)

			_, err = s.Status()
			So(err, ShouldNotBeNil)
		})

		Convey("Stopping snapd", func() {
			So(snapd, ShouldNotBeNil)

			if snapd != nil {
				snapd.Stop()
				snapd.CleanAndEraseOutput()
			}
		})
	})
}
