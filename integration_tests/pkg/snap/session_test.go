package snap

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/snap"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSnap(t *testing.T) {
	var s *snap.Session
	var publisher *wmap.PublishWorkflowMapNode
	var metricsFile string

	Convey("While having Snapteld running", t, func() {

		cleanup, loader, snapteldAddr := testhelpers.RunAndTestSnaptel()
		defer cleanup()

		Convey("We are able to connect with snapteld", func() {

			c, err := client.New(snapteldAddr, "v1", true)
			So(err, ShouldBeNil)

			Convey("Loading collectors", func() {
				err := loader.Load(snap.DockerCollector)
				So(err, ShouldBeNil)

				// Wait until metric is available in namespace.
				retries := 50
				found := false
				for i := 0; i < retries && !found; i++ {
					m := c.GetMetricCatalog()
					So(m.Err, ShouldBeNil)
					for _, metric := range m.Catalog {
						if metric.Namespace == "/intel/docker/*/stats/cgroups/cpu_stats/cpu_usage/total_usage" {
							found = true
							break
						}
					}
					time.Sleep(100 * time.Millisecond)
				}
				So(found, ShouldBeTrue)

				Convey("Loading publishers", func() {
					err := loader.Load(snap.FilePublisher)
					So(err, ShouldBeNil)

					publisher = wmap.NewPublishNode("file", snap.PluginAnyVersion)

					So(publisher, ShouldNotBeNil)

					tmpFile, err := ioutil.TempFile("", "session_test")
					tmpFile.Close()
					So(err, ShouldBeNil)

					metricsFile = tmpFile.Name()

					publisher.AddConfigItem("file", metricsFile)

					Convey("While starting a Snap experiment session", func() {
						s = snap.NewSession(
							"swan-test-dummy",
							[]string{"/intel/docker/root/stats/cgroups/cpu_stats/cpu_usage/total_usage"},
							1*time.Second,
							0,
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

							err = s.Wait()
							So(err, ShouldBeNil)
							Convey("Task is in ended state, and cannot be stopped", func() {
								err = s.Stop()
								So(err, ShouldNotBeNil)
							})

							So(s.IsRunning(), ShouldBeFalse)
						})
					})
				})
			})
		})
	})
}
