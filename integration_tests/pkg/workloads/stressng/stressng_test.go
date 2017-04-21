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
	"flag"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/stressng"
	. "github.com/smartystreets/goconvey/convey"
)

func validateExecutor(launcher executor.Launcher) func() {
	return func() {
		Convey("workload should be running", func() {
			taskHandle, err := launcher.Launch()
			if taskHandle != nil {
				defer taskHandle.Stop()
				defer taskHandle.EraseOutput()
			}
			So(err, ShouldBeNil)

			stopped := taskHandle.Wait(1 * time.Second)
			So(stopped, ShouldBeFalse)
			So(taskHandle.Status(), ShouldEqual, executor.RUNNING)

			Convey("When we stop the task", func() {
				err := taskHandle.Stop()
				Convey("There should be no error", func() {
					So(err, ShouldBeNil)
					Convey("and task should be terminated and the task status should be -1", func() {
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
}

// TestStressng  is an integration test with local executor
func TestStressng(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("While using Local Shell in stress-ng launcher", t, func() {
		l := executor.NewLocal()

		Convey("manually created", func() {
			custom := stressng.New(l, "foo", "-c 1")
			Convey("When binary is launched", validateExecutor(custom))
		})

		Convey("new stream based", func() {
			stream := stressng.NewStream(l)
			Convey("When binary is launched", validateExecutor(stream))
		})

		Convey("new l1 based", func() {
			cachel1 := stressng.NewCacheL1(l)
			Convey("When binary is launched", validateExecutor(cachel1))
		})

		Convey("new l3 based", func() {
			cachel3 := stressng.NewCacheL3(l)
			Convey("When binary is launched", validateExecutor(cachel3))
		})

		Convey("new memcpy based", func() {
			memcpy := stressng.NewMemCpy(l)
			Convey("When binary is launched", validateExecutor(memcpy))
		})

		Convey("new custom based", func() {
			flag.Set(stressng.StressngCustomArguments.Flag.Name, "-c 1")
			custom := stressng.NewCustom(l)
			Convey("When binary is launched", validateExecutor(custom))
			Reset(func() {
				flag.Set(stressng.StressngCustomArguments.Flag.Name, "")
			})
		})

	})

}
