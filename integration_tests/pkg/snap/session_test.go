package snap

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/snap"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSnap(t *testing.T) {
	var snapteld *testhelpers.Snapteld
	var s *snap.Session
	var publisher *wmap.PublishWorkflowMapNode
	var metricsFile string

	Convey("While having Snapteld running", t, func() {
		snapteld = testhelpers.NewSnapteldOnDefaultPorts()
		err := snapteld.Start()
		So(err, ShouldBeNil)

		time.Sleep(5 * time.Second)
		So(snapteld.Connected(), ShouldBeTrue)

		defer func() {
			if snapteld != nil {
				err := snapteld.Stop()
				err2 := snapteld.CleanAndEraseOutput()
				So(err, ShouldBeNil)
				So(err2, ShouldBeNil)
			}
		}()
		snapteldAddress := fmt.Sprintf("http://127.0.0.1:%d", snapteld.Port())

		Convey("We are able to connect with snapteld", func() {
			c, err := client.New(snapteldAddress, "v1", true)
			So(err, ShouldBeNil)

			loaderConfig := snap.DefaultPluginLoaderConfig()
			loaderConfig.SnapteldAddress = snapteldAddress
			pluginLoader, err := snap.NewPluginLoader(loaderConfig)
			So(err, ShouldBeNil)

			Convey("Loading collectors", func() {
				err = pluginLoader.Load(snap.DockerCollector)
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
					err := pluginLoader.Load(snap.FilePublisher)
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

							Convey("Reading samples from file", func() {

								err := s.Wait()
								So(err, ShouldBeNil)

								err := s.Stop()
								So(err, ShouldBeNil)
								So(s.IsRunning(), ShouldBeFalse)

								// one measurement should contains more then one metric.
								oneMeasurement, err := testhelpers.GetOneMeasurementFromFile(metricsFile)
								So(err, ShouldBeNil)
								So(len(oneMeasurement), ShouldBeGreaterThan, 0)

								metric, err := testhelpers.GetMetric(`/intel/docker/root/stats/cgroups/cpu_stats/cpu_usage/total_usage`, oneMeasurement)
								So(err, ShouldBeNil)
								So(metric.Tags["foo"], ShouldEqual, "bar")

								host, err := os.Hostname()
								So(err, ShouldBeNil)
								So(metric.Tags["plugin_running_on"], ShouldEqual, host)
							},
							)
						})
					})
				})
			})
		})
	})

}
