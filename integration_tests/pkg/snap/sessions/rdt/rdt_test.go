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

package rdt

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/rdt"
	. "github.com/smartystreets/goconvey/convey"
	"os"
)

func TestSnapRDTSession(t *testing.T) {

	SkipConvey("When testing RDT Session ", t, func() {
		Convey("We have snapteld running ", func() {

			cleanupSnaptel, loader, snapteldAddress := testhelpers.RunAndTestSnaptel()
			defer cleanupSnaptel()

			Convey("And we loaded publisher plugin", func() {

				tmpFile, err := ioutil.TempFile("", "rdt-session-test")
				So(err, ShouldBeNil)

				publisherMetricsFile := tmpFile.Name()
				loader.Load(snap.FilePublisher)
				tmpFile.Close()
				defer os.Remove(publisherMetricsFile)

				pluginName, _, err := snap.GetPluginNameAndType(snap.FilePublisher)
				So(err, ShouldBeNil)

				publisher := wmap.NewPublishNode(pluginName, snap.PluginAnyVersion)
				So(publisher, ShouldNotBeNil)

				publisher.AddConfigItem("file", publisherMetricsFile)

				Convey("Then we prepared and launch RDT collection session", func() {

					sessionConfig := rdt.DefaultConfig()
					sessionConfig.SnapteldAddress = snapteldAddress
					sessionConfig.Publisher = publisher
					session, err := rdt.NewSessionLauncher(sessionConfig)
					So(err, ShouldBeNil)

					tags := make(map[string]interface{})
					tags["foo"] = "bar"
					handle, err := session.LaunchSession(nil, tags)
					So(err, ShouldBeNil)

					defer func() {
						err := handle.Stop()
						So(err, ShouldBeNil)
					}()

					time.Sleep(5 * time.Second)
					Convey("Later we checked if task is running", func() {
						So(handle.Status(), ShouldEqual, executor.RUNNING)

						Convey("In order to read published data", func() {
							content, err := ioutil.ReadFile(publisherMetricsFile)
							So(err, ShouldBeNil)
							So(content, ShouldNotBeEmpty)
						})
					})
				})
			})
		})
	})
}
