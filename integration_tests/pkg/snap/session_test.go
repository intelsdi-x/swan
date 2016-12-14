// +build parallel

package snap

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
	var snapteld *testhelpers.Snapteld
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
		columns := strings.Split(lines[0], "|")
		So(len(columns), ShouldEqual, 4)
		tags := strings.Split(columns[1], ",")
		So(tags, ShouldHaveLength, 1)

		So(columns[1], ShouldEqual, "/intel/mock/foo")
		//TODO(iwan): uncomment when we upgrade Snap (0.14 does not save tags in file publisher)
		/*So("foo=bar", ShouldBeIn, tags)
		host, err := os.Hostname()
		So(err, ShouldBeNil)
		So("plugin_running_on="+host, ShouldBeIn, tags)*/
	}

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
		snapteldAddress := fmt.Sprintf("http://127.0.0.1:%d", snapteld.Port())

		Convey("We are able to connect with snapteld", func() {
			c, err := client.New(snapteldAddress, "v1", true)
			So(err, ShouldBeNil)

			loaderConfig := snap.DefaultPluginLoaderConfig()
			loaderConfig.SnapteldAddress = snapteldAddress
			pluginLoader, err := snap.NewPluginLoader(loaderConfig)
			So(err, ShouldBeNil)

			Convey("Loading collectors", func() {
				err = pluginLoader.Load(snap.MockCollector)
				So(err, ShouldBeNil)

				// Wait until metric is available in namespace.
				retries := 50
				found := false
				for i := 0; i < retries && !found; i++ {
					m := c.GetMetricCatalog()
					So(m.Err, ShouldBeNil)
					for _, metric := range m.Catalog {
						if metric.Namespace == "/intel/mock/foo" {
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

					publisher = wmap.NewPublishNode("mock-file", snap.PluginAnyVersion)

					So(publisher, ShouldNotBeNil)

					tmpFile, err := ioutil.TempFile("", "session_test")
					tmpFile.Close()
					So(err, ShouldBeNil)

					metricsFile = tmpFile.Name()

					publisher.AddConfigItem("file", metricsFile)

					Convey("While starting a Snap experiment session", func() {
						s = snap.NewSession(
							[]string{"/intel/mock/foo"},
							1*time.Second,
							c,
							publisher,
						)
						So(s, ShouldNotBeNil)
						s.CollectNodeConfigItems = append(s.CollectNodeConfigItems, snap.CollectNodeConfigItem{Ns: "/intel/mock", Key: "password", Value: "some random password"})

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
