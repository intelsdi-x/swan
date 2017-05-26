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

package mutilate

import (
	"path"
	"testing"

	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/mutilate"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSnapMutilateSession(t *testing.T) {

	Convey("When testing MutilateSnapSession ", t, func() {
		Convey("We have snapd running ", func() {

			cleanupSnap, loader, snapteldAddress := testhelpers.RunAndTestSnaptel()
			defer cleanupSnap()

			Convey("And we loaded publisher plugin", func() {

				cleanupMerticsFile, publisher, publisherDataFilePath := testhelpers.PreparePublisher(loader)
				defer cleanupMerticsFile()

				Convey("Then we prepared and launch mutilate session", func() {

					mutilateSessionConfig := mutilatesession.DefaultConfig()
					mutilateSessionConfig.SnapteldAddress = snapteldAddress
					mutilateSessionConfig.Publisher = publisher
					mutilateSnapSession, err := mutilatesession.NewSessionLauncher(mutilateSessionConfig)
					So(err, ShouldBeNil)

					cleanupMockedFile, mockedTaskInfo := testhelpers.PrepareMockedTaskInfo(path.Join(
						testhelpers.SwanPath, "plugins/snap-plugin-collector-mutilate/mutilate/mutilate.stdout"))
					defer cleanupMockedFile()

					tags := make(map[string]interface{})
					tags["foo"] = "bar"
					handle, err := mutilateSnapSession.LaunchSession(mockedTaskInfo, tags)
					So(err, ShouldBeNil)

					defer func() {
						err := handle.Stop()
						So(err, ShouldBeNil)
					}()

					Convey("Later we checked if task is running", func() {
						So(handle.Status(), ShouldEqual, executor.RUNNING)

						// These are results from test output file
						// in "src/github.com/intelsdi-x/swan/plugins/
						// snap-plugin-collector-mutilate/mutilate/mutilate.stdout"
						expectedMetrics := map[string]string{
							"avg":  "20.80000",
							"std":  "23.10000",
							"min":  "11.90000",
							"5th":  "13.30000",
							"10th": "13.40000",
							"90th": "33.40000",
							"95th": "43.10000",
							"99th": "59.50000",
							"qps":  "4993.10000",
						}

						Convey("In order to read and test published data", func() {
							dataValid := testhelpers.ReadAndTestPublisherData(publisherDataFilePath, expectedMetrics)
							So(dataValid, ShouldBeTrue)
						})
					})
				})
			})
		})
	})
}
