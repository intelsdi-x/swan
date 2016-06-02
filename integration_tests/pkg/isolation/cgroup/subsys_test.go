package integration

import (
	"os"
	"testing"

	"github.com/intelsdi-x/swan/pkg/executor"
	. "github.com/intelsdi-x/swan/pkg/isolation/cgroup"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCgroupSubsysMounts(t *testing.T) {

	// SubsysMounts()
	Convey("When reporting cgroup subsystem mounts", t, func() {
		mounts, err := SubsysMounts(executor.NewLocal(), DefaultCommandTimeout)
		Convey("The returned mounts should not be nil", func() {
			So(mounts, ShouldNotBeNil)
		})
		Convey("The returned mounts should not be empty", func() {
			So(mounts, ShouldNotBeEmpty)
		})
		Convey("The returned error should be nil", func() {
			So(err, ShouldBeNil)
		})
	})
}

// Subsys()
func TestCgroupSubsys(t *testing.T) {
	Convey("When reporting whether certain subsystems are mounted", t, func() {
		Convey("The cpu subsystem should be mounted", func() {
			mounted, err := Subsys("cpu", executor.NewLocal(), DefaultCommandTimeout)
			So(err, ShouldBeNil)
			So(mounted, ShouldBeTrue)
		})
		Convey("And the foobar subsystem should not be mounted", func() {
			mounted, err := Subsys("foobar", executor.NewLocal(), DefaultCommandTimeout)
			So(err, ShouldBeNil)
			So(mounted, ShouldBeFalse)
		})
	})
}

// SubsysPath()
func TestCgroupSubsysPath(t *testing.T) {
	Convey("When reporting the mount for a given subsys", t, func() {
		Convey("The cpu subsystem mount should exist", func() {
			mount, err := SubsysPath("cpu", executor.NewLocal(), DefaultCommandTimeout)
			So(err, ShouldBeNil)
			info, err := os.Stat(mount)
			So(err, ShouldBeNil)
			So(info, ShouldNotBeNil)
			So(info.IsDir(), ShouldBeTrue)
		})
	})
}
