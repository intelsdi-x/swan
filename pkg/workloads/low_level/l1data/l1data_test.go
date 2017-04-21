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

package l1data

import (
	"errors"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestL1dAggressor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	mockedExecutor := new(mocks.Executor)
	mockedTask := new(mocks.TaskHandle)

	Convey("While using l1d aggressor launcher", t, func() {
		const (
			pathToBinary = "test"
			validCommand = "test 86400"
		)

		Convey("Default configuration should be valid", func() {
			config := DefaultL1dConfig()
			config.Path = pathToBinary
			launcher := New(
				mockedExecutor,
				config,
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
		Convey("While using incorrect configuration", func() {
			duration := time.Duration(-1 * time.Second)
			incorrectConfiguration := Config{
				Path:     pathToBinary,
				Duration: duration,
			}
			launcher := New(mockedExecutor, incorrectConfiguration)

			Convey("Should launcher return error", func() {

				task, err := launcher.Launch()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldStartWith, "launcher configuration is invalid.")
				So(task, ShouldBeNil)
			})
		})

	})
}
