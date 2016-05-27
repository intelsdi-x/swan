package mutilate

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"path"
	"testing"
)

func TestStdoutParser(t *testing.T) {
	Convey("Opening non-existing file should fail", t, func() {
		data, err := parseOutput("/non/existing/file")

		So(data, ShouldBeEmpty)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "open /non/existing/file: no such file or directory")
	})

	Convey("Opening readable and correct file should provide meaningful results", t, func() {
		path, err := getCurrentDirFilePath("/mutilate.stdout")
		So(err, ShouldBeNil)

		data, err := parseOutput(path)

		So(err, ShouldBeNil)
		So(data, ShouldHaveLength, 10)
		So(data["avg"], ShouldResemble, 20.8)
		So(data["std"], ShouldResemble, 23.1)
		So(data["min"], ShouldResemble, 11.9)
		So(data["percentile/5th"], ShouldResemble, 13.3)
		So(data["percentile/10th"], ShouldResemble, 13.4)
		So(data["percentile/90th"], ShouldResemble, 33.4)
		So(data["percentile/95th"], ShouldResemble, 43.1)
		So(data["percentile/99th"], ShouldResemble, 59.5)
		So(data["percentile/99.999th/custom"], ShouldResemble, 1777.887805)
		So(data["qps"], ShouldResemble, 4993.1)
	})

	Convey("Attempting to read file with wrong number of read columns should return an error and no metrics", t, func() {
		path, err := getCurrentDirFilePath("/mutilate_incorrect_count_of_columns.stdout")
		So(err, ShouldBeNil)

		data, err := parseOutput(path)

		So(data, ShouldHaveLength, 0)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "Incorrect number of fields: expected 8 but got 2")
	})

	Convey("Attempting to read a file with no read row at all should return no metrics", t, func() {
		path, err := getCurrentDirFilePath("/mutilate_missing_read_row.stdout")
		So(err, ShouldBeNil)

		data, err := parseOutput(path)
		So(err, ShouldBeNil)

		// QPS and custom percentile latency are still available, thus 2.
		So(data, ShouldHaveLength, 2)

		So(data, ShouldNotContainKey, "avg")
		So(data, ShouldNotContainKey, "std")
		So(data, ShouldNotContainKey, "min")
		So(data, ShouldNotContainKey, "percentile/5th")
		So(data, ShouldNotContainKey, "percentile/10th")
		So(data, ShouldNotContainKey, "percentile/90th")
		So(data, ShouldNotContainKey, "percentile/95th")
		So(data, ShouldNotContainKey, "percentile/99th")
	})

	Convey("Attempting to read a file with no swan-specific row at all should return an error and no metrics", t, func() {
		path, err := getCurrentDirFilePath("/mutilate_missing_swan_row.stdout")
		So(err, ShouldBeNil)

		data, err := parseOutput(path)
		So(err, ShouldBeNil)

		// QPS and read latencies are still available, thus 9.
		So(data, ShouldHaveLength, 9)

		So(data, ShouldNotContainKey, "percentile/99.999th/custom")
	})

	Convey("Attempting to read a file with malformed swan-specific row should return an error and no metrics", t, func() {
		path, err := getCurrentDirFilePath("/mutilate_malformed_swan_row.stdout")
		So(err, ShouldBeNil)

		data, err := parseOutput(path)

		So(data, ShouldHaveLength, 0)
		So(err.Error(), ShouldEqual, "Incorrect number of fields: expected 2 but got 1")
	})

	Convey("Attempting to read a file with swan-specific row missing metric value should return an error and no metrics", t, func() {
		path, err := getCurrentDirFilePath("/mutilate_missing_metric_in_swan_row.stdout")
		So(err, ShouldBeNil)

		data, err := parseOutput(path)

		So(data, ShouldHaveLength, 0)
		So(err.Error(), ShouldEqual, "Incorrect number of fields: expected 2 but got 1")
	})

	Convey("Attempting to read a file with swan-specific row missing percentile value should return an error and no metrics", t, func() {
		path, err := getCurrentDirFilePath("/mutilate_swan_row_missing_percentile_in_description.stdout")
		So(err, ShouldBeNil)

		data, err := parseOutput(path)

		So(data, ShouldHaveLength, 0)
		So(err.Error(), ShouldEqual, "Incorrect number of fields: expected 2 but got 0")
	})

	Convey("Attempting to read a file with read row containing incorrect values should return an error and no results", t, func() {
		path, err := getCurrentDirFilePath("/mutilate_non_numeric_default_metric_value.stdout")
		So(err, ShouldBeNil)

		data, err := parseOutput(path)

		So(data, ShouldHaveLength, 0)
		So(err.Error(), ShouldEqual, "Incorrect number of fields: expected 8 but got 3")
	})
}

func getCurrentDirFilePath(name string) (string, error) {
	gwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return path.Join(gwd, name), nil
}
