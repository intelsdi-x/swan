package main

import (
	"github.com/intelsdi-x/snap/control"
	"github.com/intelsdi-x/snap/core"
	//	"github.com/intelsdi-x/swan/misc/snap-plugin-collector-mutilate/mutilate"
	//	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

const ()

func TestMutilatePluginLoad(t *testing.T) {
	//	snapPath := os.Getenv("SNAP_PATH")
	//	if snapPath != "" {
	// Helper plugin trigger build if possible for this plugin
	//	helper.BuildPlugin("collector", "mutilate")
	//
	Convey("Ensure mutilate plugin can be loaded", t, func() {
		pluginControl := control.New(control.GetDefaultConfig())
		pluginControl.Start()
		//		path, _ := mutilate.Get_current_dir_file("snap-plugin-collector-mutilate")
		requestedPlugin, requestedPluginError := core.NewRequestedPlugin("/home/developer/go/src/github.com/intelsdi-x/swan/misc/snap-plugin-collector-mutilate/snap-plugin-collector-mutilate")
		So(requestedPluginError, ShouldBeNil)

		//		fmt.Printf("\nrequestedPlugin: %v", requestedPlugin)
		//		fmt.Printf("\nerror: %v", requestedPluginError)
		_, loadError := pluginControl.Load(requestedPlugin)
		So(loadError, ShouldBeNil)
	})
}

//else
//		t.Skip("SNAP_PATH environme not set. Cannot test mutilate plugin.\n")
//	}
//}

func TestMutilatePluginLaunch(t *testing.T) {
	Convey("Ensure mutilate plugin can be launched", t, func() {
		os.Args = []string{"", "{\"NoDaemon\": true}"}
		So(func() { main() }, ShouldNotPanic)
	})
}
