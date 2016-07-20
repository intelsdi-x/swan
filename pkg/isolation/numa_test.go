package isolation

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNumaDecorator(t *testing.T) {
    Convey("When I want to use numa decorator", t, func() {
        Convey("Without any additional parameters", func() {
            decorator := NewNuma(false, false, []int{}, []int{}, []int{}, []int{}, -1)
            So(decorator.Decorate("test"), ShouldEqual, "numactl  -- test")
        })
        Convey("Which runs test command w/o CPU awareness", func() {
            decorator := NewNuma(true, false, []int{}, []int{}, []int{}, []int{}, -1)
            So(decorator.Decorate("test"), ShouldEqual, "numactl -a -- test")
        })
        Convey("Which is allocating memory for test command on valid nodes", func() {
            decorator := NewNuma(false, true, []int{1, 3, 4}, []int{}, []int{}, []int{}, -1)
            So(decorator.Decorate("test"), ShouldEqual, "numactl -l -i 1,3,4 -- test")
        })
        Convey("Which is allocating memory for test command on non-valid nodes", func() {
            decorator := NewNuma(false, false, []int{1,-3,6}, []int{}, []int{}, []int{}, -1)
            So(decorator.Decorate("test"), ShouldEqual, "numactl -i 1,6 -- test")
        })
        Convey("With every possible parameter", func() {
            decorator := NewNuma(false, false, []int{4}, []int{1}, []int{2}, []int{3}, 3)
            So(decorator.Decorate("test"), ShouldEqual, "numactl -i 4 -m 1 -C 3 -N 2 --preffered=3 -- test")
        })
    })

}
