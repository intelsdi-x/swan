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

func TestRdtsetDecorator(t *testing.T) {
	Convey("When I want to use rdtset decorator", t, func() {
		Convey("It should parse configuration values as expected", func() {
			decorator := &Rdtset{Mask: 2047, CPURange: "0-3"}
			command := decorator.Decorate("ls -l")

			So(command, ShouldEqual, "rdtset -v -c 0-3 -t 'l3=0x7ff;cpu=0-3' ls -l")
		})
	})

}
