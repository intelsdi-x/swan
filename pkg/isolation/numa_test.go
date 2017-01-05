package isolation

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNumaDecorator(t *testing.T) {
	Convey("When I want to use numa decorator", t, func() {
		Convey("Without any additional parameters", func() {
			decorator := Numactl{PreferredNode: -1}
			So(decorator.Decorate("test"), ShouldEqual, "numactl  -- test")
		})
		Convey("Which runs test command w/o CPU awareness", func() {
			decorator := Numactl{IsAll: true, PreferredNode: -1}
			So(decorator.Decorate("test"), ShouldEqual, "numactl --all -- test")
		})
		Convey("Which is allocating memory for test command on valid nodes", func() {
			decorator := Numactl{IsLocalalloc: true, InterleaveNodes: []int{1, 3, 4}}
			So(decorator.Decorate("test"), ShouldEqual, "numactl --localalloc --interleave=1,3,4 -- test")
		})
		Convey("Which is allocating memory for test command on non-valid nodes", func() {
			decorator := Numactl{InterleaveNodes: []int{1, -3, 6}, PreferredNode: -1}
			So(decorator.Decorate("test"), ShouldEqual, "numactl --interleave=1,6 -- test")
		})
		Convey("With every possible parameter", func() {
			decorator := Numactl{InterleaveNodes: []int{4}, MembindNodes: []int{1}, CPUnodebindNodes: []int{2}, PhyscpubindCPUs: []int{3}, PreferredNode: 3}
			So(decorator.Decorate("test"), ShouldEqual, "numactl --interleave=4 --membind=1 --physcpubind=3 --cpunodebind=2 --preferred=3 -- test")
		})
	})

}
