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
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/swan/pkg/workloads/caffe"
)

func TestCaffeWithMockedExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("When I create Caffe with local executor and default configuration", t, func() {
		localExecutor := executor.NewLocal()
		c := caffe.New(localExecutor, caffe.DefaultConfig())

		Convey("When I launch the workload", func() {
			handle, err := c.Launch()
			defer handle.Stop()
			defer handle.EraseOutput()

			Convey("Error is nil", func() {
				So(err, ShouldBeNil)

				Convey("Proper handle is returned", func() {
					So(handle, ShouldNotBeNil)

					Convey("Should work for at least one sec", func() {
						isTerminated, err := handle.Wait(1 * time.Second)
						So(err, ShouldBeNil)
						So(isTerminated, ShouldBeFalse)

						Convey("Should be able to stop with no problem and be terminated", func() {
							err = handle.Stop()
							So(err, ShouldBeNil)

							state := handle.Status()
							So(state, ShouldEqual, executor.TERMINATED)
						})
					})
				})
			})
		})
	})
}
