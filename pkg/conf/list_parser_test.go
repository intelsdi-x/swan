package conf

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/alecthomas/kingpin.v2"
)

func TestStringListValue(t *testing.T) {
	Convey("While using Custom StringListValue parser", t, func() {
		strListValue := StringListValue([]string{})

		Convey("It should implement kinping.Value interfaces", func() {
			So(strListValue, ShouldImplement, (*kingpin.Value)(nil))
			So(strListValue, ShouldImplement, (*kingpin.Getter)(nil))
		})

		Convey("When parsing string inputs it should append them to string slice", func() {
			So(strListValue.IsCumulative(), ShouldBeTrue)

			So(strListValue.Set("A"), ShouldBeNil)
			So(strListValue.Get(), ShouldResemble, []string{"A"})

			So(strListValue.Set("B"), ShouldBeNil)
			So(strListValue.Get(), ShouldResemble, []string{"A", "B"})

			So(strListValue.Set("C,D"), ShouldBeNil)
			So(strListValue.Get(), ShouldResemble, []string{"A", "B", "C", "D"})

			So(strListValue.String(), ShouldEqual, strings.Join([]string{"A", "B", "C", "D"}, ","))
		})
	})
}
