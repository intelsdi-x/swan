package isolation

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSet(t *testing.T) {
	Convey("When testing sets", t, func() {

		Convey("When creating an empty set", func() {
			s1 := NewSet()

			Convey("Should have length 0", func() {
				So(s1, ShouldHaveLength, 0)
			})
		})

		Convey("Creating a set with three elements", func() {
			s1 := NewSet(1, 5, 7)

			Convey("Should have length 3", func() {
				So(s1, ShouldHaveLength, 3)

				Convey("And the elements should be present", func() {
					_, ok := s1[1]
					So(ok, ShouldBeTrue)

					_, ok = s1[5]
					So(ok, ShouldBeTrue)

					_, ok = s1[7]
					So(ok, ShouldBeTrue)
				})
			})
		})

		Convey("Creating a set with from range string \"0-5,34,46-48\"", func() {
			s1, err := NewSetFromRange("0-5,34,46-48")
			So(err, ShouldBeNil)

			Convey("Should have length 10", func() {
				So(s1, ShouldHaveLength, 10)

				Convey("And the elements should be present", func() {
					_, ok := s1[0]
					So(ok, ShouldBeTrue)

					_, ok = s1[1]
					So(ok, ShouldBeTrue)

					_, ok = s1[2]
					So(ok, ShouldBeTrue)

					_, ok = s1[3]
					So(ok, ShouldBeTrue)

					_, ok = s1[4]
					So(ok, ShouldBeTrue)

					_, ok = s1[5]
					So(ok, ShouldBeTrue)

					_, ok = s1[34]
					So(ok, ShouldBeTrue)

					_, ok = s1[46]
					So(ok, ShouldBeTrue)

					_, ok = s1[47]
					So(ok, ShouldBeTrue)

					_, ok = s1[48]
					So(ok, ShouldBeTrue)
				})
			})
		})
	})
}
