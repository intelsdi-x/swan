package mutilate

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/exp/inotify"
	"os"
	"testing"
)

func TestStdoutParser(t *testing.T) {
	Convey("Opening non-existing file should fail", t, func() {
		event := inotify.Event{Name: "/non/existing/file"}

		data, error := parseMutilateStdout(event)

		So(data, ShouldBeZeroValue)
		So(error.Error(), ShouldEqual,
			"open /non/existing/file: no such file or directory")

	})

	Convey("Opening non-readable file should fail", t, func() {
		event := inotify.Event{Name: "/etc/shadow"}

		data, error := parseMutilateStdout(event)

		So(data, ShouldBeZeroValue)
		So(error.Error(), ShouldEqual, "open /etc/shadow: permission denied")
	})

	Convey("Opening readable and correct file should provide meaningful results",
		t, func() {
			event := inotify.Event{Name: GetCurrentDirFilePath("/mutilate.stdout")}

			data, error := parseMutilateStdout(event)

			So(error, ShouldBeNil)
			So(data, ShouldHaveLength, 9)
			soMetric := func(mutilateMetric metric, name string, value float64) {
				So(mutilateMetric, ShouldResemble, metric{name, value})
			}
			soMetric(data[0], "avg", 20.8)
			soMetric(data[1], "std", 23.1)
			soMetric(data[2], "min", 11.9)
			soMetric(data[3], "percentile/5th", 13.3)
			soMetric(data[4], "percentile/10th", 13.4)
			soMetric(data[5], "percentile/90th", 33.4)
			soMetric(data[6], "percentile/95th", 43.1)
			soMetric(data[7], "percentile/99th", 59.5)
			soMetric(data[8], "percentile/99.999th", 1777.887805)
		})

	Convey("Attempting to read file with wrong number of read columns should return"+
		" an error and no metrics", t, func() {
		event := inotify.Event{Name: GetCurrentDirFilePath(
			"/mutilate_incorrect_count_of_columns.stdout")}

		data, error := parseMutilateStdout(event)

		So(data, ShouldHaveLength, 0)
		So(error.Error(), ShouldEqual,
			"Incorrect column count (got: 3, expected: 9) in QPS read row")
	})

	Convey("Attempting to read a file with no read row at all should return"+
		" an error and no metrics", t, func() {
		event := inotify.Event{Name: GetCurrentDirFilePath(
			"/mutilate_missing_read_row.stdout")}

		data, error := parseMutilateStdout(event)

		So(data, ShouldHaveLength, 0)
		So(error.Error(), ShouldEqual, "No default mutilate statistics found")
	})

	Convey("Attempting to read a file with no swan-specific row at all"+
		" should return an error and no metrics", t, func() {
		event := inotify.Event{Name: GetCurrentDirFilePath("/mutilate_missing_swan_row.stdout")}

		data, error := parseMutilateStdout(event)

		So(data, ShouldHaveLength, 0)
		So(error.Error(), ShouldEqual, "No swan-specific statistics found")
	})

	Convey("Attempting to read a file with malformed swan-specific row should return"+
		" an error and no metrics", t, func() {
		event := inotify.Event{Name: GetCurrentDirFilePath("/mutilate_malformed_swan_row.stdout")}

		data, error := parseMutilateStdout(event)

		So(data, ShouldHaveLength, 0)
		So(error.Error(), ShouldEqual, "Swan-specific row malformed")
	})

	Convey("Attempting to read a file with swan-specific row missing metric"+
		" value should return an error and no metrics", t, func() {
		event := inotify.Event{Name: GetCurrentDirFilePath("/mutilate_missing_metric_in_swan_row.stdout")}

		data, error := parseMutilateStdout(event)

		So(data, ShouldHaveLength, 0)
		So(error.Error(), ShouldEqual, "Swan-specific row is missing metric value")
	})

	Convey("Attempting to read a file with swan-specific row missing percentile value"+
		" should return an error and no metrics", t, func() {
		//		mutilate_swan_row_missing_percentile_in_description.stdout
		event := inotify.Event{Name: GetCurrentDirFilePath(
			"/mutilate_swan_row_missing_percentile_in_description.stdout")}

		data, error := parseMutilateStdout(event)

		So(data, ShouldHaveLength, 0)
		So(error.Error(), ShouldEqual, "Swan-specific row is missing percentile value")
	})

	Convey("Attempting to read a file with read row containing incorrect values should return"+
		" en error and no results", t, func() {
		event := inotify.Event{Name: GetCurrentDirFilePath(
			"/mutilate_non_numeric_default_metric_value.stdout")}

		data, error := parseMutilateStdout(event)

		So(data, ShouldHaveLength, 0)
		So(error.Error(), ShouldEqual,
			"Non-numeric value in read row (value: \"thisIsNotANumber\", column: 5)")
	})
}

func GetCurrentDirFilePath(name string) string {
	var path bytes.Buffer
	gwd, _ := os.Getwd()
	path.WriteString(gwd)
	path.WriteString(name)

	return path.String()
}
