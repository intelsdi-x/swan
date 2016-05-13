// +build integration

package main

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMutilatePluginLaunch(t *testing.T) {
	Convey("Ensure mutilate plugin can be launched", t, func() {
		os.Args = []string{"", "{\"NoDaemon\": true}"}
		So(func() { main() }, ShouldNotPanic)
	})
}

