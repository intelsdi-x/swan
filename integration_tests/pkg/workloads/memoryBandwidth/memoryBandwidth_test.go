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

package integration

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/memoryBandwidth"
	. "github.com/smartystreets/goconvey/convey"
)

// TestMemBwDataWithExecutor is an integration test with local executor
func TestMemBwDataWithExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("While using Local Shell in Memory Bandwidth launcher", t, func() {
		l := executor.NewLocal()
		memBwDataLauncher := memoryBandwidth.New(
			l, memoryBandwidth.DefaultMemBwConfig())

		Convey("When memBwd binary is launched", func() {
			taskHandle, err := memBwDataLauncher.Launch()
			if taskHandle != nil {
				defer taskHandle.Stop()
				defer taskHandle.EraseOutput()
			}

			Convey("There should be no error", func() {
				stopErr := taskHandle.Stop()

				So(err, ShouldBeNil)
				So(stopErr, ShouldBeNil)
			})

			Convey("MemBwData should be running", func() {
				So(taskHandle.Status(), ShouldEqual, executor.RUNNING)
			})

			Convey("When we stop the memBw task", func() {
				err := taskHandle.Stop()
				Convey("There should be no error", func() {
					So(err, ShouldBeNil)
				})
				Convey("The task should be terminated and the task status should be -1", func() {
					taskState := taskHandle.Status()
					So(taskState, ShouldEqual, executor.TERMINATED)

					exitCode, err := taskHandle.ExitCode()

					So(err, ShouldBeNil)
					So(exitCode, ShouldEqual, -1)
				})
			})
		})

	})

}
