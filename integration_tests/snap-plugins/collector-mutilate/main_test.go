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
		requestedPlugin, requestedPluginError := core.NewRequestedPlugin(pluginPath)
		So(requestedPluginError, ShouldBeNil)

		_, loadError := pluginControl.Load(requestedPlugin)
		So(loadError, ShouldBeNil)
	})
}
