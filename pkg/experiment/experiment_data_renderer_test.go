package experiment

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/cassandra"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestExperimentDataRenderer(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)
	tagsMap := map[string]string{"swan_experiment": "fakeID", "swan_phase": "fakeID", "swan_repetition": "0"}
	Convey("While getting metric for double type", t, func() {
		doubleval := 10.000000
		metrics := cassandra.NewMetrics("fakeID", 1, "abc", time.Now(), false, doubleval, "c", tagsMap, "doubleval")
		Convey("I should receive string value of given doubleval", func() {
			So(getStringFromMetricValue(metrics.Valtype(), metrics), ShouldEqual, "10.000000")
		})
	})
	Convey("While getting metric for string type", t, func() {
		stringval := "abc"
		metrics := cassandra.NewMetrics("fakeID", 1, "abc", time.Now(), false, 10, stringval, tagsMap, "strval")
		Convey("I should receive string value of given string", func() {
			So(getStringFromMetricValue(metrics.Valtype(), metrics), ShouldEqual, "abc")
		})
	})
	Convey("While getting metric for boolval type", t, func() {
		boolval := true
		metrics := cassandra.NewMetrics("fakeID", 1, "abc", time.Now(), boolval, 10, "c", tagsMap, "boolval")
		Convey("I should receive string value of given bool", func() {
			So(getStringFromMetricValue(metrics.Valtype(), metrics), ShouldEqual, "true")
		})
	})
	Convey("While checking value in slice", t, func() {
		sliceOfValues := []string{"a", "b"}
		Convey("I should receive true if it exists", func() {
			So(isValueInSlice("a", sliceOfValues), ShouldEqual, true)
		})
		Convey("I should receive false if it does not exist", func() {
			So(isValueInSlice("c", sliceOfValues), ShouldEqual, false)
		})
	})
	Convey("While creating list of unique values from map", t, func() {
		sliceOfValues := []string{"1", "2"}
		Convey("If key is present in map", func() {
			Convey("and value is not in list, I should receive a list with this value", func() {
				elem := map[string]string{"a": "3"}
				returnedList := createUniqueList("a", elem, sliceOfValues)
				So(len(returnedList), ShouldEqual, 1)
				So(returnedList[0], ShouldEqual, "3")
			})
			Convey("and value is in list, I should receive empty list", func() {
				elem := map[string]string{"a": "1"}
				returnedList := createUniqueList("a", elem, sliceOfValues)
				So(len(returnedList), ShouldEqual, 0)
			})
		})
		Convey("If key is not present in map", func() {
			Convey("I should receive empty list", func() {
				elem := map[string]string{"a": "1"}
				returnedList := createUniqueList("c", elem, sliceOfValues)
				So(len(returnedList), ShouldEqual, 0)
			})
		})

	})
}
