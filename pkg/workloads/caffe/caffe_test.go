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

package caffe

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestCaffeWithMockedExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("When I create Caffe with mocked executor and default configuration", t, func() {
		mExecutor := new(executor.MockExecutor)
		mHandle := new(executor.MockTaskHandle)

		c := New(mExecutor, DefaultConfig())
		Convey("When I launch the workload with success", func() {
			mExecutor.On("Execute", mock.AnythingOfType("string")).Return(mHandle, nil).Once()
			handle, err := c.Launch()
			Convey("Proper handle is returned", func() {
				So(handle, ShouldEqual, mHandle)
			})
			Convey("Error is nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When I launch the workload with failure", func() {
			expectedErr := errors.New(`"caffe.sh test -model examples/cifar10/cifar10_quick_train_test.prototxt -weights examples/cifar10/cifar10_quick_iter_5000.caffemodel.h5 -iterations 1000000000 -sigint_effect stop"`)
			mExecutor.On("Execute", mock.AnythingOfType("string")).Return(nil, expectedErr).Once()
			handle, err := c.Launch()
			Convey("Proper handle is returned", func() {
				So(handle, ShouldBeNil)
			})
			Convey("Error is not nil and root cause is passed", func() {
				So(err.Error(), ShouldContainSubstring, expectedErr.Error())
			})
		})
	})
}

func TestCaffeDefaultConfig(t *testing.T) {
	Convey("When I create default config for Caffe", t, func() {
		config := DefaultConfig()
		Convey("Binary field is not blank", func() {
			So(config.BinaryPath, ShouldNotBeBlank)
		})
	})
}
