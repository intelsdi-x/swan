package cgroup

import (
	"testing"

	"github.com/intelsdi-x/swan/pkg/isolation"
	. "github.com/smartystreets/goconvey/convey"
)

// NewCgroup(controllers []string, path string)
func TestNewCgroup(t *testing.T) {
	Convey("When properly constructing a cgroup", t, func() {
		controllers := []string{"cpu", "cpuset"}
		path := "foo"
		cg, err := NewCgroup(controllers, path)
		Convey("The returned cgroup should not be nil", func() {
			So(cg, ShouldNotBeNil)
		})
		Convey("It should implement isolation.Isolation", func() {
			So(cg, ShouldImplement, (*isolation.Isolation)(nil))
		})
		Convey("The returned error should be nil", func() {
			So(err, ShouldBeNil)
		})
	})
	Convey("When improperly constructing a cgroup with empty controllers", t, func() {
		cg, err := NewCgroup([]string{}, "foo")
		Convey("The returned cgroup should be nil", func() {
			So(cg, ShouldBeNil)
		})
		Convey("And the returned error should not be nil", func() {
			So(err, ShouldNotBeNil)
		})
	})
	Convey("When improperly constructing a cgroup with an empty path", t, func() {
		cg, err := NewCgroup([]string{"foo", "bar"}, "")
		Convey("The returned cgroup should be nil", func() {
			So(cg, ShouldBeNil)
		})
		Convey("And the returned error should not be nil", func() {
			So(err, ShouldNotBeNil)
		})
	})
}

// Controllers() []string
func TestCgroupControllers(t *testing.T) {
	Convey("After constructing a cgroup", t, func() {
		controllers := []string{"cpuset"}
		cg, _ := NewCgroup(controllers, "foo")
		Convey("It should have the right controllers", func() {
			So(cg.Controllers(), ShouldResemble, controllers)
		})
	})
}

// Path() string
func TestCgroupPath(t *testing.T) {
	Convey("After constructing a cgroup", t, func() {
		path := "foo"
		cg, _ := NewCgroup([]string{"cpuset"}, path)
		Convey("It should have the right path", func() {
			So(cg.Path(), ShouldEqual, "/foo")
		})
	})
}

// IsRoot() string
func TestCgroupIsRoot(t *testing.T) {
	Convey("After constructing a top-level cgroup", t, func() {
		path := "foo"
		cg, _ := NewCgroup([]string{"cpuset"}, path)

		Convey("It should not be the root cgroup", func() {
			So(cg.IsRoot(), ShouldBeFalse)
		})
		Convey("Its parent should be the root cgroup", func() {
			So(cg.Parent().IsRoot(), ShouldBeTrue)
		})
	})
}

// Ancestors() string
func TestCgroupAncestors(t *testing.T) {
	Convey("After constructing a nested cgroup", t, func() {
		path := "foo/bar/baz"
		cg, _ := NewCgroup([]string{"cpuset"}, path)

		Convey("There should be exactly three cgroups in the ancestor list", func() {
			So(len(cg.Ancestors()), ShouldEqual, 3)
		})

		Convey("And the first ancestor should be the root of the cgroup hierarchy", func() {
			So(cg.Ancestors()[0].IsRoot(), ShouldBeTrue)
		})

		Convey("And the last ancestor should have the same path as the original's parent", func() {
			So(cg.Ancestors()[2].Path(), ShouldEqual, cg.Parent().Path())
		})
	})
}
