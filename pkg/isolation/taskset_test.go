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

package isolation

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTasksetDecorator(t *testing.T) {
	Convey("When I want to use taskset decorator", t, func() {
		Convey("With simple one cpu range", func() {
			decorator := Taskset{NewIntSet(1)}
			So(decorator.Decorate("test"), ShouldEqual, "taskset -c 1 test")
		})

		Convey("With simple complex cpu range", func() {
			decorator := Taskset{NewIntSet(1, 3, 4, 7, 8)}
			So(decorator.Decorate("test"), ShouldEqual, "taskset -c 1,3,4,7,8 test")
		})
	})

}
