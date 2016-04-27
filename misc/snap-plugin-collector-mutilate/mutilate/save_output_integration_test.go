package mutilate

import (
	"bytes"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/exp/inotify"
	"os"
	"testing"
	"time"
)

func TestOutputParser(t *testing.T) {
	Convey("Opening non-existing file should fail", t, func() {
		event := inotify.Event{Name: "/non/existing/file"}
		baseTime := time.Now()

		data, error := parse_mutilate_save_output(event, baseTime)

		So(data, ShouldBeZeroValue)
		So(error.Error(), ShouldEqual, "open /non/existing/file: no such file or directory")

	})

	Convey("Opening non-readable file should fail", t, func() {
		event := inotify.Event{Name: "/etc/shadow"}
		baseTime := time.Now()

		data, error := parse_mutilate_save_output(event, baseTime)

		So(data, ShouldBeZeroValue)
		So(error.Error(), ShouldEqual, "open /etc/shadow: permission denied")
	})

	Convey("Opening readable and correct file should provide meaningful results", t, func() {
		event := inotify.Event{Name: get_current_dir_file("/mutilate.out")}
		cest, _ := time.LoadLocation("Europe/Warsaw")
		baseTime := time.Date(2010, time.April, 10, 8, 41, 0, 0, cest)
		zeroTime := time.Date(2010, time.April, 10, 8, 10, 26, 0, cest)

		data, error := parse_mutilate_save_output(event, baseTime)

		So(error, ShouldBeNil)
		So(data, ShouldHaveLength, 10)
		So(data[0].latency, ShouldEqual, 70.095062)
		So(data[0].time.Unix(), ShouldEqual, zeroTime.Unix())
		So(data[1].latency, ShouldEqual, 28.848648)
		So(data[1].time.Unix(), ShouldEqual, zeroTime.Unix())
		So(data[2].latency, ShouldEqual, 14.781952)
		So(data[2].time.Unix(), ShouldEqual, zeroTime.Unix())
		So(data[3].latency, ShouldEqual, 14.066696)
		So(data[3].time.Unix(), ShouldEqual, zeroTime.Unix())
		So(data[4].latency, ShouldEqual, 14.066696)
		So(data[4].time.Unix(), ShouldEqual, zeroTime.Unix())
		So(data[5].latency, ShouldEqual, 14.066696)
		So(data[5].time.Unix(), ShouldEqual, zeroTime.Unix())
		So(data[6].latency, ShouldEqual, 12.874603)
		So(data[6].time.Unix(), ShouldEqual, zeroTime.Unix())
		So(data[7].latency, ShouldEqual, 12.874603)
		So(data[7].time.Unix(), ShouldEqual, zeroTime.Unix())
		So(data[8].latency, ShouldEqual, 14.066696)
		So(data[8].time.Unix(), ShouldEqual, zeroTime.Unix())
		So(data[9].latency, ShouldEqual, 13.828278)
		So(data[9].time.Unix(), ShouldEqual, baseTime.Unix())
	})
}

func TestGettingFirstRowTime(t *testing.T) {
	Convey("We should get correct timestamp for first row", t, func() {
		file, _ := os.Open(get_current_dir_file("/mutilate.out"))
		defer file.Close()
		cest, _ := time.LoadLocation("Europe/Warsaw")
		baseTime := time.Date(2010, time.April, 10, 8, 41, 0, 0, cest)
		expectedFirstRowTime := time.Date(2010, time.April, 10, 8, 10, 26, 0, cest)

		realFirstRowTime := get_first_save_row_time(file, baseTime)
		fmt.Printf("\n%v\n%v", expectedFirstRowTime, realFirstRowTime)

		So(realFirstRowTime.Equal(expectedFirstRowTime), ShouldBeTrue)
	})
}

func get_current_dir_file(name string) string {
	var path bytes.Buffer
	gwd, _ := os.Getwd()
	path.WriteString(gwd)
	path.WriteString(name)

	return path.String()
}
