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
	"os/exec"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	stream "github.com/intelsdi-x/swan/pkg/workloads/low_level/stream"
	. "github.com/smartystreets/goconvey/convey"
)

// TestStreamWithExecutor is an integration test with local executor.
func TestStreamWithExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	if _, err := exec.LookPath("stream.100M"); err != nil {
		t.Skip("stream.100M binary not found")
	}

	Convey("While using Local Shell in stream launcher", t, func() {
		l := executor.NewLocal()
		streamLauncher := stream.New(l, stream.DefaultConfig())

		Convey("When stream binary is launched", func() {
			taskHandle, err := streamLauncher.Launch()
			Convey("task should launch successfully", func() {
				So(err, ShouldBeNil)
				Reset(func() {
					taskHandle.Stop()
					taskHandle.EraseOutput()
				})
				Convey("and stream should be running", func() {
					So(taskHandle.Status(), ShouldEqual, executor.RUNNING)
					Convey("When we stop the stream task", func() {
						err := taskHandle.Stop()
						Convey("There should be no error", func() {
							So(err, ShouldBeNil)
							Convey("and task should be terminated and the task exit status should be -1 (killed)", func() {
								taskState := taskHandle.Status()
								So(taskState, ShouldEqual, executor.TERMINATED)

								exitCode, err := taskHandle.ExitCode()

								So(err, ShouldBeNil)
								So(exitCode, ShouldEqual, -1)
							})
						})
					})
				})
			})

		})

	})

}
