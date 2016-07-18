package sysctl

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSysctl(t *testing.T) {
	Convey("Reading a sane sysctl value (kernel.ostype)", t, func() {
		value, err := Get("kernel.ostype")

		Convey("Should not return an error", func() {
			So(err, ShouldBeNil)

			Convey("And should contain 'Linux'", func() {
				So(value, ShouldEqual, "Linux")
			})
		})
	})

	Convey("Reading a non-sense sysctl value (foo.bar.baz)", t, func() {
		value, err := Get("foo.bar.baz")

		Convey("Should return an error", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "open /proc/sys/foo/bar/baz: no such file or directory")

			Convey("And the value should be empty", func() {
				So(value, ShouldEqual, "")
			})
		})
	})
}
