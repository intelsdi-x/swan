package isolation

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTasksetDecorator(t *testing.T) {
	Convey("When I want to use taskset decorator", t, func() {
		Convey("I can pass valid non-empty list of cpus my Decorator should returns decorator with cpus ", func() {
			decorator := NewTasksetDecorator([]int{0, 3, 4})
			So(decorator.Decorate("test"), ShouldEqual, "taskset -c 0,3,4 test")
		})

		Convey("I can pass empty list of cpus my Decorator should returns decorator which is pinning command into cpu #0 ", func() {
			decorator := NewTasksetDecorator([]int{})
			So(decorator.Decorate("test"), ShouldEqual, "taskset -c 0 test")
		})

		Convey("I can pass half-valid list of cpus my Decorator should returns decorator which is pinning command into valid cpus", func() {
			decorator := NewTasksetDecorator([]int{0, -4, -5, 7, 3})
			So(decorator.Decorate("test"), ShouldEqual, "taskset -c 0,7,3 test")
		})

		Convey("I can pass non-valid list of cpus my Decorator should returns decorator which is pinning command into cpu #0", func() {
			decorator := NewTasksetDecorator([]int{-1, -4, -5, -8, -3})
			So(decorator.Decorate("test"), ShouldEqual, "taskset -c 0 test")
		})

	})
}
