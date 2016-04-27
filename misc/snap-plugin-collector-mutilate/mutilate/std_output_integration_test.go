package mutilate

import (
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/exp/inotify"
	"testing"
	"time"
)

func TestStdoutParser(t *testing.T) {
	Convey("Opening non-existing file should fail", t, func() {
		event := inotify.Event{Name: "/non/existing/file"}
		baseTime := time.Now()

		data, error := parse_mutilate_stdout(event, baseTime)

		So(data, ShouldBeZeroValue)
		So(error.Error(), ShouldEqual, "open /non/existing/file: no such file or directory")

	})

	Convey("Opening non-readable file should fail", t, func() {
		event := inotify.Event{Name: "/etc/shadow"}
		baseTime := time.Now()

		data, error := parse_mutilate_stdout(event, baseTime)

		So(data, ShouldBeZeroValue)
		So(error.Error(), ShouldEqual, "open /etc/shadow: permission denied")
	})

	Convey("Opening readable and correct file should provide meaningful results", t, func() {
		event := inotify.Event{Name: get_current_dir_file("/mutilate.stdout")}
		baseTime := time.Now()

		data, error := parse_mutilate_stdout(event, baseTime)

		So(error, ShouldBeNil)
		So(data, ShouldHaveLength, 9)
		So(data[0].name, ShouldEqual, "avg")
		So(data[0].value, ShouldEqual, 20.8)
		So(data[0].time.Unix(), ShouldEqual, baseTime.Unix())
		So(data[1].name, ShouldEqual, "std")
		So(data[1].value, ShouldEqual, 23.1)
		So(data[1].time.Unix(), ShouldEqual, baseTime.Unix())
		So(data[2].name, ShouldEqual, "min")
		So(data[2].value, ShouldEqual, 11.9)
		So(data[2].time.Unix(), ShouldEqual, baseTime.Unix())
		So(data[3].name, ShouldEqual, "percentile/5th")
		So(data[3].value, ShouldEqual, 13.3)
		So(data[3].time.Unix(), ShouldEqual, baseTime.Unix())
		So(data[4].name, ShouldEqual, "percentile/10th")
		So(data[4].value, ShouldEqual, 13.4)
		So(data[4].time.Unix(), ShouldEqual, baseTime.Unix())
		So(data[5].name, ShouldEqual, "percentile/90th")
		So(data[5].value, ShouldEqual, 33.4)
		So(data[5].time.Unix(), ShouldEqual, baseTime.Unix())
		So(data[6].name, ShouldEqual, "percentile/95th")
		So(data[6].value, ShouldEqual, 43.1)
		So(data[6].time.Unix(), ShouldEqual, baseTime.Unix())
		So(data[7].name, ShouldEqual, "percentile/99th")
		So(data[7].value, ShouldEqual, 59.5)
		So(data[7].time.Unix(), ShouldEqual, baseTime.Unix())
		So(data[8].name, ShouldEqual, "percentile/99.999th")
		So(data[8].value, ShouldEqual, 1777.887805)
		So(data[8].time.Unix(), ShouldEqual, baseTime.Unix())

	})
}
