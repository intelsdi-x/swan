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

			Convey("Should be empty", func() {
				So(s1.Empty(), ShouldBeTrue)
			})
		})

		Convey("Creating a set with three elements", func() {
			s1 := NewSet(1, 5, 7)

			Convey("Should not be empty", func() {
				So(s1.Empty(), ShouldBeFalse)
			})

			Convey("Should have length 3", func() {
				So(s1, ShouldHaveLength, 3)

				Convey("And the elements should be present", func() {
					So(s1.Contains(1), ShouldBeTrue)
					So(s1.Contains(5), ShouldBeTrue)
					So(s1.Contains(7), ShouldBeTrue)
				})

				Convey("And other elements should not be present", func() {
					So(s1.Contains(2), ShouldBeFalse)
					So(s1.Contains(4), ShouldBeFalse)
					So(s1.Contains(6), ShouldBeFalse)
				})
			})
		})

		Convey("Creating two sets that share only one element", func() {
			s1 := NewSet(1, 3, 5, 7)
			s2 := NewSet(2, 4, 6, 7)

			Convey("Each set should equal itself", func() {
				So(s1.Equals(s1), ShouldBeTrue)
				So(s2.Equals(s2), ShouldBeTrue)

				Convey("But they should not equal each other", func() {
					So(s1.Equals(s2), ShouldBeFalse)
					So(s2.Equals(s1), ShouldBeFalse)
				})
			})

			Convey("The difference should not include the shared element", func() {
				diff12 := s1.Difference(s2)
				diff21 := s2.Difference(s1)
				So(diff12.Contains(7), ShouldBeFalse)
				So(diff21.Contains(7), ShouldBeFalse)

				Convey("But they should contain all other original elements", func() {
					So(diff12.Contains(1), ShouldBeTrue)
					So(diff12.Contains(3), ShouldBeTrue)
					So(diff12.Contains(5), ShouldBeTrue)

					So(diff21.Contains(2), ShouldBeTrue)
					So(diff21.Contains(4), ShouldBeTrue)
					So(diff21.Contains(6), ShouldBeTrue)
				})

				Convey("And difference with themselves should be the empty set", func() {
					So(s1.Difference(s1).Empty(), ShouldBeTrue)
					So(s2.Difference(s2).Empty(), ShouldBeTrue)
				})
			})

			Convey("Their union should contain all items", func() {
				union := s1.Union(s2)
				So(len(union), ShouldEqual, 7)
				So(union.Contains(1), ShouldBeTrue)
				So(union.Contains(2), ShouldBeTrue)
				So(union.Contains(3), ShouldBeTrue)
				So(union.Contains(4), ShouldBeTrue)
				So(union.Contains(5), ShouldBeTrue)
				So(union.Contains(6), ShouldBeTrue)
				So(union.Contains(7), ShouldBeTrue)

				Convey("And union with themselves should equal the original", func() {
					So(s1.Union(s1).Equals(s1), ShouldBeTrue)
					So(s2.Union(s2).Equals(s2), ShouldBeTrue)
				})
			})

			Convey("Their intersection should contain only the shared item", func() {
				intersection := s1.Intersection(s2)
				So(len(intersection), ShouldEqual, 1)
				So(intersection.Contains(7), ShouldBeTrue)
			})
		})

		Convey("Creating a set with from range string \"0-5,34,46-48\"", func() {
			s1, err := NewSetFromRange("0-5,34,46-48")
			So(err, ShouldBeNil)

			Convey("Should have length 10", func() {
				So(s1, ShouldHaveLength, 10)

				Convey("And the elements should be present", func() {
					So(s1.Contains(0), ShouldBeTrue)
					So(s1.Contains(1), ShouldBeTrue)
					So(s1.Contains(2), ShouldBeTrue)
					So(s1.Contains(3), ShouldBeTrue)
					So(s1.Contains(4), ShouldBeTrue)
					So(s1.Contains(5), ShouldBeTrue)

					So(s1.Contains(34), ShouldBeTrue)

					So(s1.Contains(46), ShouldBeTrue)
					So(s1.Contains(47), ShouldBeTrue)
					So(s1.Contains(48), ShouldBeTrue)
				})

				Convey("And other elements should not be present", func() {
					So(s1.Contains(-2), ShouldBeFalse)
					So(s1.Contains(32), ShouldBeFalse)
					So(s1.Contains(50), ShouldBeFalse)
				})
			})
		})
	})
}
