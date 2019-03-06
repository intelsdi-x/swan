// Copyright (c) 2019 Intel Corporation
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

package redis

import (
	"errors"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/isolation"
	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"syscall"
	"testing"
	"time"
)

// IsEndpointListeningMockedSuccess is a mocked Is
func IsEndpointListeningMockedSuccess(address string, timeout time.Duration) bool {
	return true
}

func IsEndpointListeningMockedFailure(address string, timeout time.Duration) bool {
	return false
}

func TestRedisWithMockedExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	const (
		expectedCommand = "test --port 6379 --bind 0.0.0.0 --maxmemory 512mb --protected-mode no"
		expectedHost    = "127.0.0.1"
	)

	Convey("When I create PID namespace isolation", t, func() {
		mockedExecutor := new(executor.MockExecutor)
		mockedTaskHandle := new(executor.MockTaskHandle)
		var decorators []isolation.Decorator
		unshare, err := isolation.NewNamespace(syscall.CLONE_NEWPID)
		So(err, ShouldBeNil)
		decorators = append(decorators, unshare)

		Convey("While using Redis launcher", func() {

			config := DefaultConfig()
			config.PathToBinary = "test"

			redisLauncher := New(
				mockedExecutor,
				config)

			redisLauncher.isRedisUp = IsEndpointListeningMockedSuccess

			Convey("While simulating proper execution", func() {
				mockedExecutor.On("Execute", expectedCommand).Return(mockedTaskHandle, nil).Once()
				mockedTaskHandle.On("Address").Return(expectedHost)

				Convey("Build command should create proper command", func() {
					command := redisLauncher.buildCommand()
					So(command, ShouldEqual, expectedCommand)

					Convey("Arguments passed to Executor should be a proper command", func() {
						task, err := redisLauncher.Launch()
						So(err, ShouldBeNil)

						So(task, ShouldNotBeNil)
						So(task, ShouldEqual, mockedTaskHandle)

						Convey("Location of the returned task shall be 127.0.0.1", func() {
							addr := task.Address()
							So(addr, ShouldEqual, expectedHost)
							mockedTaskHandle.AssertExpectations(t)
						})
						mockedExecutor.AssertExpectations(t)
					})

					Convey("When test connection to redis fails task handle shall be nil and error shall be return", func() {
						mockedTaskHandle.On("Stop").Return(nil)
						mockedTaskHandle.On("Clean").Return(nil)
						mockedTaskHandle.On("EraseOutput").Return(nil)
						redisLauncher.isRedisUp = IsEndpointListeningMockedFailure
						task, err := redisLauncher.Launch()
						So(err, ShouldNotBeNil)
						So(task, ShouldBeNil)

						mockedExecutor.AssertExpectations(t)
					})

					Convey("When test connection to redis fails and task.Stop fails task handle shall be nil and error shall be return", func() {
						mockedTaskHandle.On("Stop").Return(errors.New("Test error code for stop"))
						mockedTaskHandle.On("Clean").Return(nil)
						mockedTaskHandle.On("EraseOutput").Return(nil)
						redisLauncher.isRedisUp = IsEndpointListeningMockedFailure
						task, err := redisLauncher.Launch()
						So(err, ShouldNotBeNil)
						So(task, ShouldBeNil)

						mockedExecutor.AssertExpectations(t)
					})

					Convey("When test connection to memcached fails, task.Stop and task.EraseOutput succeeds but task.Clean fails task handle shall be nil and error shall be return", func() {
						mockedTaskHandle.On("Stop").Return(nil)
						mockedTaskHandle.On("Clean").Return(errors.New("Test error code for clean"))
						mockedTaskHandle.On("EraseOutput").Return(nil)
						redisLauncher.isRedisUp = IsEndpointListeningMockedFailure
						task, err := redisLauncher.Launch()
						So(err, ShouldNotBeNil)
						So(task, ShouldBeNil)

						mockedExecutor.AssertExpectations(t)
					})

					Convey("When test connection to memcached fails, task.Stop and task.Clean succeeds but task.EraseOutput fails task handle shall be nil and error shall be return", func() {
						mockedTaskHandle.On("Stop").Return(nil)
						mockedTaskHandle.On("Clean").Return(nil)
						mockedTaskHandle.On("EraseOutput").Return(errors.New("Test error code for erasing output"))
						redisLauncher.isRedisUp = IsEndpointListeningMockedFailure
						task, err := redisLauncher.Launch()
						So(err, ShouldNotBeNil)
						So(task, ShouldBeNil)

						mockedExecutor.AssertExpectations(t)
					})

				})

			})

			Convey("While simulating error execution", func() {
				mockedExecutor.On("Execute", expectedCommand).Return(nil, errors.New("test")).Once()

				Convey("Build command should create proper command", func() {
					command := redisLauncher.buildCommand()
					So(command, ShouldEqual, expectedCommand)

					Convey("Arguments passed to Executor should be a proper command", func() {
						task, err := redisLauncher.Launch()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "test")

						So(task, ShouldBeNil)

						mockedExecutor.AssertExpectations(t)
					})
				})

			})

		})

	})
}
