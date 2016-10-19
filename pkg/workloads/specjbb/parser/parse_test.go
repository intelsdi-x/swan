package parser

import (
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStdoutParser(t *testing.T) {
	Convey("Opening non-existing file for hbir rt should fail", t, func() {
		jops, err := FileWithHBIRRT("/non/existing/file")

		Convey("jops should equal 0 and the error should not be nil", func() {
			So(jops, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "open /non/existing/file: no such file or directory")
		})
	})

	Convey("Opening readable and correct file for hbir rt", t, func() {
		path, err := filepath.Abs("criticaljops")
		So(err, ShouldBeNil)

		Convey("should provide meaningful results", func() {
			jops, err := FileWithHBIRRT(path)
			So(err, ShouldBeNil)
			So(jops, ShouldEqual, 2684)
		})
	})

	Convey("Attempting to read file without measured critical jops", t, func() {
		path, err := filepath.Abs("criticaljops_not_measured")
		So(err, ShouldBeNil)
		Convey("should return 0 value for jops and an error", func() {
			jops, err := FileWithHBIRRT(path)
			So(jops, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Incorrect number of fields: expected 1 but got 0")
		})
	})

	Convey("Opening non-existing file for latencies should fail", t, func() {
		results, err := FileWithLatencies("/non/existing/file")

		Convey("results should be nil and the error should not be nil", func() {
			So(results.Raw, ShouldHaveLength, 0)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "open /non/existing/file: no such file or directory")
		})
	})

	Convey("Opening readable and correct file for latencies", t, func() {
		path, err := filepath.Abs("latencies")
		So(err, ShouldBeNil)

		Convey("should provide meaningful results", func() {
			results, err := FileWithLatencies(path)
			So(err, ShouldBeNil)
			So(results.Raw, ShouldHaveLength, 12)
			So(results.Raw["Success"], ShouldEqual, 115276)
			So(results.Raw["Partial"], ShouldEqual, 0)
			So(results.Raw["Failed"], ShouldEqual, 0)
			So(results.Raw["SkipFail"], ShouldEqual, 0)
			So(results.Raw["Probes"], ShouldEqual, 108937)
			So(results.Raw["Samples"], ShouldEqual, 133427)
			So(results.Raw["min"], ShouldEqual, 300)
			So(results.Raw["percentile/50th"], ShouldEqual, 3100)
			So(results.Raw["percentile/90th"], ShouldEqual, 21000)
			So(results.Raw["percentile/95th"], ShouldEqual, 89000)
			So(results.Raw["percentile/99th"], ShouldEqual, 517000)
			So(results.Raw["max"], ShouldEqual, 640000)
		})
	})

	Convey("Attempting to read file without measured latencies", t, func() {
		path, err := filepath.Abs("latencies_not_measured")
		So(err, ShouldBeNil)
		Convey("should return 0 results and an error", func() {
			results, err := FileWithLatencies(path)
			So(results.Raw, ShouldHaveLength, 0)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Incorrect number of fields: expected 12 but got 0")
		})
	})
	Convey("Opening non-existing file for raw file name should fail", t, func() {
		fileName, err := FileWithRawFileName("/non/existing/file")

		Convey("file name should be nil and the error should not be nil", func() {
			So(fileName, ShouldEqual, "")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "open /non/existing/file: no such file or directory")
		})
	})

	Convey("Opening readable and correct file for raw file name", t, func() {
		path, err := filepath.Abs("raw_file_name")
		So(err, ShouldBeNil)

		Convey("should provide meaningful results", func() {
			fileName, err := FileWithRawFileName(path)
			So(err, ShouldBeNil)
			So(fileName, ShouldEqual, "/swan/workloads/web_serving/specjbb/specjbb2015-D-20160921-00002.data.gz")
		})
	})

	Convey("Attempting to read file without raw file name", t, func() {
		path, err := filepath.Abs("raw_file_name_not_given")
		So(err, ShouldBeNil)
		Convey("should result in an empty name and an error", func() {
			fileName, err := FileWithRawFileName(path)
			So(fileName, ShouldEqual, "")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Raw file name not found")
		})
	})
}
