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

package sysctl

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSysctl(t *testing.T) {
	Convey("Reading a sane sysctl value (kernel.ostype)", t, func() {
		value, err := Get("kernel.ostype")

		Convey("Should not return an error", func() {
			So(err, ShouldBeNil)

			Convey("And should contain 'Linux'", func() {
				So(value, ShouldEqual, "Linux")
			})
		})
	})

	Convey("Reading a non-sense sysctl value (foo.bar.baz)", t, func() {
		value, err := Get("foo.bar.baz")

		Convey("Should return an error", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "open /proc/sys/foo/bar/baz: no such file or directory")

			Convey("And the value should be empty", func() {
				So(value, ShouldEqual, "")
			})
		})
	})
}
