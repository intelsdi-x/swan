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

package specjbb

import (
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

const (
	testLoad     = 60
	testDuration = 10 * time.Millisecond
)

func TestSPECjbbLoadGenerator(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("When creating load generator", t, func() {
		controller := new(executor.MockExecutor)
		transactionInjector := new(executor.MockExecutor)
		config := DefaultLoadGeneratorConfig()
		config.PathToBinary = "test"

		loadGenerator := NewLoadGenerator(controller, []executor.Executor{
			transactionInjector,
		}, config)

		Convey("And generating load", func() {
			controller.On("Execute", mock.AnythingOfType("string")).Return(new(executor.MockTaskHandle), nil)

			transactionInjector.On("Execute", mock.AnythingOfType("string")).Return(new(executor.MockTaskHandle), nil)

			loadGeneratorTaskHandle, err := loadGenerator.Load(testLoad, testDuration)

			Convey("On success, error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("On success, task handle should not be nil", func() {
				So(loadGeneratorTaskHandle, ShouldNotBeNil)
			})
		})
	})

}
