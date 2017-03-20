package isolation

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTasksetDecorator(t *testing.T) {
	Convey("When I want to use taskset decorator", t, func() {
		Convey("With simple one cpu range", func() {
			decorator := Taskset{NewIntSet(1)}
			So(decorator.Decorate("test"), ShouldEqual, "taskset --cpu-list=1 -- test")
		})

		Convey("With simple complex cpu range", func() {
			decorator := Taskset{NewIntSet(1, 3, 4, 7, 8)}
			So(decorator.Decorate("test"), ShouldEqual, "taskset --cpu-list=1,3,4,7,8 -- test")
		})
	})

}
