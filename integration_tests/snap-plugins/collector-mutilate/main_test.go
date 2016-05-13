package mutilate

import (
	"os"
	"path"
	"testing"

	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/core"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMutilatePluginLoad(t *testing.T) {
	// TODO(niklas): Fix race (https://intelsdi.atlassian.net/browse/SCE-316)
	SkipConvey("Ensure mutilate plugin can be loaded", t, func() {
		basePath := os.Getenv("GOPATH")
		pluginPath := path.Join(basePath, "src", "github.com", "intelsdi-x", "swan", "misc",
			"snap-plugin-collector-mutilate", "snap-plugin-collector-mutilate")
		pluginControl := control.New(control.GetDefaultConfig())
		pluginControl.Start()
		requestedPlugin, requestedPluginError := core.NewRequestedPlugin(pluginPath)
		So(requestedPluginError, ShouldBeNil)

		_, loadError := pluginControl.Load(requestedPlugin)
		So(loadError, ShouldBeNil)
	})
}
