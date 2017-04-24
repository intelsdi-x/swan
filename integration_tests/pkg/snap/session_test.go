// Copyright (c) 2017 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
							c,
							publisher,
						)
						So(s, ShouldNotBeNil)

						tags := make(map[string]interface{})
						tags["foo"] = "bar"
						err := s.Start(tags)

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

							err = s.Stop()
							So(err, ShouldBeNil)
							So(s.IsRunning(), ShouldBeFalse)
						})
					})
				})
			})
		})
	})
}
