package snap

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/intelsdi-x/athena/integration_tests/test_helpers"
	"github.com/intelsdi-x/athena/pkg/snap"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSnap(t *testing.T) {
	t.Skipf("skipping test for now as snap plugins are not bundled with Athena")

	var snapd *testhelpers.Snapd
	var s *snap.Session
	var publisher *wmap.PublishWorkflowMapNode
	var metricsFile string
	testStopping := func() {
		s.Wait()
		err := s.Stop()
		So(err, ShouldBeNil)
		So(s.IsRunning(), ShouldBeFalse)
		dat, err := ioutil.ReadFile(metricsFile)
		So(err, ShouldBeNil)
		So(dat, ShouldNotBeEmpty)
		dat = bytes.Trim(dat, "\n\r")

		lines := strings.Split(string(dat), "\n")
		So(len(lines), ShouldEqual, 1)
		columns := strings.Split(lines[0], "\t")
		So(len(columns), ShouldEqual, 3)
		tags := strings.Split(columns[1], ",")
		So(tags, ShouldHaveLength, 1)

		So(columns[0], ShouldEqual, "/intel/swan/session/metric1")
		So("foo=bar", ShouldBeIn, tags)
		host, err := os.Hostname()
		So(err, ShouldBeNil)
		So("plugin_running_on="+host, ShouldBeIn, tags)
	}

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

		Convey("We are able to connect with snapd", func() {
			c, err := client.New(snapdAddress, "v1", true)
			So(err, ShouldBeNil)

			loaderConfig := snap.DefaultPluginLoaderConfig()
			loaderConfig.SnapdAddress = snapdAddress
			pluginLoader, err := snap.NewPluginLoader(loaderConfig)
			So(err, ShouldBeNil)

			Convey("Loading collectors", func() {
				err = pluginLoader.Load(snap.SessionCollector)
				So(err, ShouldBeNil)

				// Wait until metric is available in namespace.
				retries := 50
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
					time.Sleep(100 * time.Millisecond)
				}
				So(found, ShouldBeTrue)

				Convey("Loading publishers", func() {
					pluginLoader.Load(snap.SessionPublisher)
					So(err, ShouldBeNil)

					publisher = wmap.NewPublishNode("session-test", 1)

					So(publisher, ShouldNotBeNil)

					tmpFile, err := ioutil.TempFile("", "session_test")
					tmpFile.Close()
					So(err, ShouldBeNil)

					metricsFile = tmpFile.Name()

					publisher.AddConfigItem("file", metricsFile)

					Convey("While starting a Snap experiment session", func() {
						s = snap.NewSession(
							[]string{"/intel/swan/session/metric1"},
							1*time.Second,
							c,
							publisher,
						)
						So(s, ShouldNotBeNil)

						err := s.Start("foo:bar")

						So(err, ShouldBeNil)

						defer func() {
							if s.IsRunning() {
								err := s.Stop()
								So(err, ShouldBeNil)
							}
						}()
						Convey("Contacting snap to get the task status", func() {
							So(s.IsRunning(), ShouldBeTrue)

							Convey("Reading samples from file", testStopping)
						})
					})
				})
			})
		})
	})

}
