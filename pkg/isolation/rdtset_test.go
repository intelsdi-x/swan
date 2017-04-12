package isolation

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRdtsetDecorator(t *testing.T) {
	Convey("When I want to use rdtset decorator", t, func() {
		Convey("It should parse configuration values as expected", func() {
			decorator := &Rdtset{Mask: 2047, CPURange: "0-3"}
			command := decorator.Decorate("ls -l")

			So(command, ShouldEqual, "rdtset -v -c 0-3 -t 'l3=0x7ff;cpu=0-3' ls -l")
		})
	})

}
