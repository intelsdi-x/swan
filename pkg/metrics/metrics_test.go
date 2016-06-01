package metrics

import (
	"testing"

	"github.com/nu7hatch/gouuid"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSwanMetrics(t *testing.T) {
	tagUUID, _ := uuid.NewV4()
	phaseName := "SimplePhaseName"
	repetition := 123

	Convey("When I want to fulfill my Swan object", t, func() {
		Convey("I need to prepare a Tags structure", func() {
			tags := &Tags{
				ExperimentID: tagUUID.String(),
				PhaseID:      phaseName,
				RepetitionID: repetition,
			}
			So(tags, ShouldNotBeNil)

			Convey("Which should be comperable to other Tags instance with same values in fields", func() {
				sameTags := &Tags{
					ExperimentID: tagUUID.String(),
					PhaseID:      phaseName,
					RepetitionID: repetition,
				}
				So(tags.Compare(*sameTags), ShouldBeTrue)
				So(tags, ShouldResemble, tags)

				Convey("And it shouldn't be comperable with some other different Tags instance", func() {
					otherUUID, _ := uuid.NewV4()
					newTags := &Tags{
						ExperimentID: otherUUID.String(),
						PhaseID:      "Other task",
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
