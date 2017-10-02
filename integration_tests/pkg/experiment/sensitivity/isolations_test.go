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

package sensitivity

import (
	"testing"

	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetWorkloadCPUThreads(t *testing.T) {
	Convey("When using Automatic Core Placement CPU threads list shall not be empty", t, func() {
		hpThreads, beL1Threads, beLLCThreads := sensitivity.GetWorkloadCPUThreads()
		So(len(hpThreads), ShouldBeGreaterThan, 0)
		So(len(beL1Threads), ShouldBeGreaterThan, 0)
		So(len(beLLCThreads), ShouldBeGreaterThan, 0)
	})
}
