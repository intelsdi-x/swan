package dummy

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDummy(t *testing.T) {
	Convey("Creating a dummy instance", t, func() {
		d := NewDummy()

		Convey("Should provide foo() to return 42", func() {
			So(d.Foo(), ShouldEqual, 42)
		})
	})
}
