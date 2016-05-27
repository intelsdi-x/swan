package isolation

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCPUSelect(t *testing.T) {

	ex := map[int]int{}

	threadset := ThreadSet{requestedThreads: ex, requestID: "experiment1"}

	Convey("Should provide CPUSelect() should not return nil", t, func() {
		So(threadset.CPUSelect(0, ShareLLCButNotL1L2), ShouldNotBeNil)

		Convey("Should have length 0", func() {
			So(threadset.requestedThreads, ShouldHaveLength, 0)
		})

	})

	Convey("Should provide CPUSelect() to return nil and correct cpu ids", t, func() {
		So(threadset.CPUSelect(1, ShareLLCButNotL1L2), ShouldBeNil)

		Convey("Should have length 1", func() {
			So(threadset.requestedThreads, ShouldHaveLength, 1)
		})

	})

	Convey("Should provide CPUSelect() to return nil and correct cpu ids", t, func() {
		So(threadset.CPUSelect(2, ShareLLCButNotL1L2), ShouldBeNil)

		Convey("Should have length 2", func() {
			So(threadset.requestedThreads, ShouldHaveLength, 2)
		})

	})

	Convey("Should provide CPUSelect() to return nil and correct cpu ids", t, func() {
		So(threadset.CPUSelect(3, ShareLLCButNotL1L2), ShouldBeNil)

		Convey("Should have length 3", func() {
			So(threadset.requestedThreads, ShouldHaveLength, 3)
		})

	})

	Convey("Should provide CPUSelect() to return nil and correct cpu ids", t, func() {
		So(threadset.CPUSelect(4, ShareLLCButNotL1L2), ShouldBeNil)

		Convey("Should have length 4", func() {
			So(threadset.requestedThreads, ShouldHaveLength, 4)
		})

	})
	Convey("Should provide CPUSelect() to return nil and correct cpu ids", t, func() {
		So(threadset.CPUSelect(5, ShareLLCButNotL1L2), ShouldBeNil)

		Convey("Should have length 5", func() {
			So(threadset.requestedThreads, ShouldHaveLength, 5)
		})

	})
	Convey("Should provide CPUSelect() to return nil and correct cpu ids", t, func() {
		So(threadset.CPUSelect(6, ShareLLCButNotL1L2), ShouldBeNil)

		Convey("Should have length 6", func() {
			So(threadset.requestedThreads, ShouldHaveLength, 6)
		})

	})
	Convey("Should provide CPUSelect() to return nil and correct cpu ids", t, func() {
		So(threadset.CPUSelect(7, ShareLLCButNotL1L2), ShouldBeNil)

		Convey("Should have length 7", func() {
			So(threadset.requestedThreads, ShouldHaveLength, 7)
		})

	})
	Convey("Should provide CPUSelect() to return nil and correct cpu ids", t, func() {
		So(threadset.CPUSelect(8, ShareLLCButNotL1L2), ShouldBeNil)

		Convey("Should have length 8", func() {
			So(threadset.requestedThreads, ShouldHaveLength, 8)
		})

	})
	Convey("Should provide CPUSelect() to return nil and correct cpu ids", t, func() {
		So(threadset.CPUSelect(9, ShareLLCButNotL1L2), ShouldBeNil)

		Convey("Should have length 9", func() {
			So(threadset.requestedThreads, ShouldHaveLength, 9)
		})

	})
	Convey("Should provide CPUSelect() to return nil and correct cpu ids", t, func() {
		So(threadset.CPUSelect(10, ShareLLCButNotL1L2), ShouldBeNil)

		Convey("Should have length 10", func() {
			So(threadset.requestedThreads, ShouldHaveLength, 10)
		})

	})
}
