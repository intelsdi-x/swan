package metrics

import (
	"testing"

	"github.com/nu7hatch/gouuid"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSwanMetrics(t *testing.T) {
	tagUUID, _ := uuid.NewV4()
	phaseName := "SimplePhaseName"
	loadPoint := 2
	repetition := 123

	Convey("When I want to fulfill my SwanMetrics object", t, func() {
		Convey("I need to prepare a Tags structure", func() {
			tags := &Tags{
				ExperimentID: tagUUID.String(),
				PhaseID:      phaseName,
				LoadPoint:    loadPoint,
				RepetitionID: repetition,
			}
			So(tags, ShouldNotBeNil)

			Convey("Which should be comperable to itself", func() {
				sameTags := &Tags{
					ExperimentID: tagUUID.String(),
					PhaseID:      phaseName,
					LoadPoint:    loadPoint,
					RepetitionID: repetition,
				}
				So(tags.Compare(*sameTags), ShouldBeTrue)
				So(tags, ShouldResemble, tags)

				Convey("And it shouldn't be the same with some other Tag", func() {
					otherUUID, _ := uuid.NewV4()
					newTags := &Tags{
						ExperimentID: otherUUID.String(),
						PhaseID:      "Other task",
						LoadPoint:    10,
						RepetitionID: 321,
					}
					So(tags.Compare(*newTags), ShouldBeFalse)
					So(tags, ShouldNotResemble, newTags)
				})
			})

			Convey("And some extra metrics, packed into Metrics object", func() {
				metrics := Metadata{
					LCName: "test",
				}

				Convey("Which are used to construct SwanMetrics object", func() {
					swanMetrics := New(*tags, metrics)
					So(swanMetrics, ShouldNotBeNil)
					So(swanMetrics.Tags.Compare(*tags), ShouldBeTrue)
					So(swanMetrics.Metrics.LCName, ShouldEqual, metrics.LCName)
				})

			})

		})
	})

}
