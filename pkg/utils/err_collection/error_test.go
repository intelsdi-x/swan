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

package errcollection

import (
	"errors"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestErrorCollection(t *testing.T) {
	Convey("When use ErrorCollection", t, func() {
		var errCollection ErrorCollection

		Convey("When no error was passed, GetErr should return nil", func() {
			So(errCollection.GetErrIfAny(), ShouldBeNil)
		})
		Convey("When nil error was passed, GetErr should return nil", func() {
			errCollection.Add(nil)
			So(errCollection.GetErrIfAny(), ShouldBeNil)
		})

		Convey("When we pass one error, GetErr should return error with exact message", func() {
			errCollection.Add(errors.New("test error"))
			So(errCollection.GetErrIfAny(), ShouldNotBeNil)
			So(errCollection.GetErrIfAny().Error(), ShouldEqual, "test error")
		})

		Convey("When we pass multiple errors, GetErr should return error with combined messages", func() {
			errCollection.Add(errors.New("test error"))
			errCollection.Add(errors.New("test error1"))
			errCollection.Add(errors.New("test error2"))
			So(errCollection.GetErrIfAny(), ShouldNotBeNil)
			So(errCollection.GetErrIfAny().Error(), ShouldEqual, "test error;\n test error1;\n test error2")
		})
	})
}
