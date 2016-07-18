package topo

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewThread(t *testing.T) {
	Convey("When creating a new thread", t, func() {
		t := NewThread(1, 2, 3)

		Convey("It should have the right thread ID", func() {
			So(t.ID(), ShouldEqual, 1)
		})

		Convey("It should have the right core ID", func() {
			So(t.Core(), ShouldEqual, 2)
		})

		Convey("It should have the right socket ID", func() {
			So(t.Socket(), ShouldEqual, 3)
		})

		Convey("It should implement equality", func() {
			t2 := NewThread(1, 2, 3)
			t3 := NewThread(0, 2, 3)
			t4 := NewThread(1, 0, 3)
			t5 := NewThread(1, 2, 0)

			So(t.Equals(t2), ShouldBeTrue)
			So(t.Equals(t3), ShouldBeFalse)
			So(t.Equals(t4), ShouldBeFalse)
			So(t.Equals(t5), ShouldBeFalse)
		})
	})
}
