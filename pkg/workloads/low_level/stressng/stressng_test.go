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

package stressng

import (
	"errors"
	"testing"

	"github.com/intelsdi-x/swan/pkg/executor"
	. "github.com/smartystreets/goconvey/convey"
)

func TestStressng(t *testing.T) {

	mockedExecutor := new(executor.MockExecutor)
	mockedTask := new(executor.MockTaskHandle)

	Convey("While using stress-ng aggressor launcher", t, func() {

		Convey("Default configuration should be valid", func() {
			const validCommand = "stress-ng -foo --bar"
			launcher := New(
				mockedExecutor,
				"stress-ng",
				"-foo --bar",
			)

			Convey("When executor is able to run this command then it should return mocked taskHandle "+
				"without error", func() {

				mockedExecutor.On("Execute", validCommand).Return(mockedTask, nil).Once()

				task, err := launcher.Launch()
				So(err, ShouldBeNil)
				So(task, ShouldEqual, mockedTask)

				mockedExecutor.AssertExpectations(t)
			})
			Convey("When executor isn't able to run this command then it should return error without "+
				"mocked taskHandle", func() {

				mockedExecutor.On("Execute", validCommand).Return(nil, errors.New("fail to execute")).Once()

				task, err := launcher.Launch()
				So(task, ShouldBeNil)
				So(err.Error(), ShouldEqual, "fail to execute")

				mockedExecutor.AssertExpectations(t)
			})
		})

		Convey("and for specific aggressors we got command as expected", func() {

			Convey("for new stream based aggressor", func() {
				launcher := NewStream(mockedExecutor)
				So(launcher.String(), ShouldEqual, "stress-ng-stream")
				mockedExecutor.On("Execute", "stress-ng --stream=1").Return(mockedTask, nil).Once()
				_, err := launcher.Launch()
				So(err, ShouldBeNil)
				mockedExecutor.AssertExpectations(t)

			})

			Convey("for new l1 intensive aggressor", func() {
				launcher := NewCacheL1(mockedExecutor)
				So(launcher.String(), ShouldEqual, "stress-ng-cache-l1")
				mockedExecutor.On("Execute", "stress-ng --cache=1 --cache-level=1").Return(mockedTask, nil).Once()
				_, err := launcher.Launch()
				So(err, ShouldBeNil)
				mockedExecutor.AssertExpectations(t)

			})

			Convey("for new l3 intensive aggressor", func() {
				launcher := NewCacheL3(mockedExecutor)
				So(launcher.String(), ShouldEqual, "stress-ng-cache-l3")
				mockedExecutor.On("Execute", "stress-ng --cache=1 --cache-level=3").Return(mockedTask, nil).Once()
				_, err := launcher.Launch()
				So(err, ShouldBeNil)
				mockedExecutor.AssertExpectations(t)

			})

			Convey("for new memcpy aggressor", func() {
				launcher := NewMemCpy(mockedExecutor)
				So(launcher.String(), ShouldEqual, "stress-ng-memcpy")
				mockedExecutor.On("Execute", "stress-ng --memcpy=1").Return(mockedTask, nil).Once()
				_, err := launcher.Launch()
				So(err, ShouldBeNil)
				mockedExecutor.AssertExpectations(t)

			})

			Convey("for new custom aggressor", func() {
				launcher := NewCustom(mockedExecutor)
				So(launcher.String(), ShouldEqual, "stress-ng-custom ")
				mockedExecutor.On("Execute", "stress-ng ").Return(mockedTask, nil).Once()
				_, err := launcher.Launch()
				So(err, ShouldBeNil)
				mockedExecutor.AssertExpectations(t)

			})
		})

	})
}
