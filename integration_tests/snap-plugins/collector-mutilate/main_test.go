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

package mutilate

import (
	"testing"

	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMutilatePluginLoad(t *testing.T) {

	pluginPath := testhelpers.AssertFileExists("snap-plugin-collector-mutilate")

	Convey("Ensure mutilate plugin can be loaded", t, func() {

		pluginControl := control.New(control.GetDefaultConfig())
		pluginControl.Start()
		requestedPlugin, requestedPluginError := core.NewRequestedPlugin(pluginPath, "/tmp/", nil)
		So(requestedPluginError, ShouldBeNil)

		_, loadError := pluginControl.Load(requestedPlugin)
		So(loadError, ShouldBeNil)
	})
}
