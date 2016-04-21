package dummy

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDummy(t *testing.T) {
	Convey("Creating a dummy instance", t, func() {
		d := NewDummy()

		Convey("Should provide foo() to return 42", func() {
			So(d.Foo(), ShouldEqual, 42)
		})
	})
}
