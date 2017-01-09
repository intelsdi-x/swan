package isolation

import (
	"fmt"
	"sort"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIntSet(t *testing.T) {
	Convey("When testing sets", t, func() {

		Convey("When creating an empty set", func() {
			s1 := NewIntSet()

			Convey("Should have length 0", func() {
				So(s1, ShouldHaveLength, 0)
			})

			Convey("Should be empty", func() {
				So(s1.Empty(), ShouldBeTrue)
			})
		})

		Convey("After adding elements to a set", func() {
			s := NewIntSet()
			numItems := 1024
			for i := 0; i < numItems; i++ {
				s.Add(i)
			}

			Convey(fmt.Sprintf("It should have length %d", numItems), func() {
				So(s, ShouldHaveLength, numItems)
			})

			Convey("It should contain all added elements", func() {
				for i := 0; i < numItems; i++ {
					So(s.Contains(i), ShouldBeTrue)
				}
			})
		})

		Convey("When removing elements from a set", func() {
			s := NewIntSet()
			numItems := 1024
			for i := 0; i < numItems; i++ {
				s.Add(i)
			}

			Convey(fmt.Sprintf("After adding %d elements", numItems), func() {
				Convey(fmt.Sprintf("It should have length %d", numItems), func() {
					So(s, ShouldHaveLength, numItems)
				})

				Convey(fmt.Sprintf("After removing %d elements", numItems/2), func() {
					for i := 0; i < numItems; i++ {
						if i%2 == 0 {
							s.Remove(i)
						}
					}

					Convey(fmt.Sprintf("It should have length %d", numItems/2), func() {
						So(s, ShouldHaveLength, numItems/2)
					})

					Convey("It should not contain any removed elements", func() {
						for i := 0; i < numItems; i++ {
							if i%2 == 0 {
								So(s.Contains(i), ShouldBeFalse)
							}
						}
					})

					Convey("It should contain all expected remaining elements", func() {
						for i := 0; i < numItems; i++ {
							if i%2 != 0 {
								So(s.Contains(i), ShouldBeTrue)
							}
						}
					})
				})
			})
		})

		Convey("Creating a set with three elements", func() {
			s1 := NewIntSet(1, 5, 7)

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
			s1 := NewIntSet(1, 3, 5, 7)
			s2 := NewIntSet(2, 4, 6, 7)

			Convey("Each set should equal itself", func() {
				So(s1.Equals(s1), ShouldBeTrue)
				So(s2.Equals(s2), ShouldBeTrue)

				Convey("But they should not equal each other", func() {
					So(s1.Equals(s2), ShouldBeFalse)
					So(s2.Equals(s1), ShouldBeFalse)
				})

				Convey("And they should not be subsets of each other", func() {
					So(s1.Subset(s2), ShouldBeFalse)
					So(s2.Subset(s1), ShouldBeFalse)
				})

				Convey("But they should be subsets of themselves", func() {
					So(s1.Subset(s1), ShouldBeTrue)
					So(s2.Subset(s2), ShouldBeTrue)
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

			Convey("The union of difference and intersection should equal the original", func() {
				So(s1.Difference(s2).Union(s1.Intersection(s2)).Equals(s1), ShouldBeTrue)
				So(s2.Difference(s1).Union(s2.Intersection(s1)).Equals(s2), ShouldBeTrue)
			})

			Convey("The intersection should be a subset of both", func() {
				So(s1.Intersection(s2).Subset(s1), ShouldBeTrue)
				So(s2.Intersection(s1).Subset(s2), ShouldBeTrue)
			})
		})

		Convey("IntSet correctly yields a subset of a given size with take", func() {
			s := NewIntSet(1, 2, 3, 4, 5, 6, 7, 8)

			r, err := s.Take(0)
			So(err, ShouldBeNil)
			So(r.Empty(), ShouldBeTrue)

			r, err = s.Take(4)
			So(err, ShouldBeNil)
			So(len(r), ShouldEqual, 4)
			So(r.Subset(s), ShouldBeTrue)
		})

		Convey("IntSet correctly converts to a slice of integers", func() {
			s1 := NewIntSet()
			s2 := NewIntSet(1, 3, 5, 7)

			So(s1.AsSlice(), ShouldResemble, []int{})

			s2slice := s2.AsSlice()
			sort.Ints(s2slice)
			So(len(s2slice), ShouldEqual, len(s2))
			So(s2slice, ShouldResemble, []int{1, 3, 5, 7})
			for _, elem := range s2slice {
				So(s2.Contains(elem), ShouldBeTrue)
			}
		})

		Convey("Creating a set with from range string \"0-5,34,46-48\"", func() {
			s1, err := NewIntSetFromRange("0-5,34,46-48")
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

		Convey("IntSet correctly converts to a range string", func() {
			s1 := NewIntSet()
			s2 := NewIntSet(1, 3, 5, 7)

			So(s1.AsRangeString(), ShouldEqual, "")
			parsed, err := NewIntSetFromRange(s1.AsRangeString())
			So(err, ShouldBeNil)
			So(parsed.Equals(s1), ShouldBeTrue)

			// Test against the result string's length because there
			// is no ordering guarantee. The result must be some permutation
			// of 1,2,5,7 but always seven characters long.
			So(len(s2.AsRangeString()), ShouldEqual, 7)
			parsed, err = NewIntSetFromRange(s2.AsRangeString())
			So(err, ShouldBeNil)
			So(parsed.Equals(s2), ShouldBeTrue)
		})
	})
}
