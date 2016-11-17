package parser

import (
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStdoutParser(t *testing.T) {
	Convey("Opening non-existing file for high bound injection rate (critical jops determined in tuning phase) should fail",
		t, func() {
			jops, err := FileWithHBIRRT("/non/existing/file")

			Convey("jops should equal 0 and the error should not be nil", func() {
				So(jops, ShouldEqual, 0)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "open /non/existing/file: no such file or directory")
			})
		})

	Convey("Opening readable and correct file for high bound injection rate (critical jops determined in tuning phase)",
		t, func() {
			path, err := filepath.Abs("criticaljops")
			So(err, ShouldBeNil)

			Convey("should provide meaningful results", func() {
				jops, err := FileWithHBIRRT(path)
				So(err, ShouldBeNil)
				So(jops, ShouldEqual, 2684)
			})
		})

	Convey("Opening readable and correct file for high bound injection rate (critical jops determined in tuning phase) from remote output",
		t, func() {
			path, err := filepath.Abs("remote_output")
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
			So(err.Error(), ShouldEqual, "Run result not found, cannot determine critical-jops")
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
			So(results.Raw, ShouldHaveLength, 14)
			So(results.Raw[SuccessKey], ShouldEqual, 115276)
			So(results.Raw[PartialKey], ShouldEqual, 0)
			So(results.Raw[FailedKey], ShouldEqual, 0)
			So(results.Raw[SkipFailKey], ShouldEqual, 0)
			So(results.Raw[ProbesKey], ShouldEqual, 108937)
			So(results.Raw[SamplesKey], ShouldEqual, 133427)
			So(results.Raw[MinKey], ShouldEqual, 300)
			So(results.Raw[Percentile50Key], ShouldEqual, 3100)
			So(results.Raw[Percentile90Key], ShouldEqual, 21000)
			So(results.Raw[Percentile95Key], ShouldEqual, 89000)
			So(results.Raw[Percentile99Key], ShouldEqual, 517000)
			So(results.Raw[MaxKey], ShouldEqual, 640000)
			So(results.Raw[QPSKey], ShouldEqual, 4007)
			So(results.Raw[IssuedRequestsKey], ShouldEqual, 4007)
		})
	})

	Convey("Attempting to read file without measured latencies", t, func() {
		path, err := filepath.Abs("latencies_not_measured")
		So(err, ShouldBeNil)
		Convey("should return 0 results and an error", func() {
			results, err := FileWithLatencies(path)
			So(results.Raw, ShouldHaveLength, 0)
			So(err, ShouldNotBeNil)
		})
	})
	Convey("Attempting to read file without measured processed requests and issued requests", t, func() {
		path, err := filepath.Abs("pr_not_measured")
		So(err, ShouldBeNil)
		Convey("should return empty results and an error", func() {
			results, err := FileWithLatencies(path)
			So(results.Raw, ShouldHaveLength, 0)
			So(err, ShouldNotBeNil)
		})
	})
	Convey("Attempting to read correct and readable file with many iterations for load", t, func() {
		path, err := filepath.Abs("many_iterations")
		So(err, ShouldBeNil)
		Convey("should return last iteration results and no error", func() {
			results, err := FileWithLatencies(path)
			So(results.Raw, ShouldHaveLength, 14)
			So(err, ShouldBeNil)
			So(results.Raw[SuccessKey], ShouldEqual, 114968)
			So(results.Raw[PartialKey], ShouldEqual, 0)
			So(results.Raw[FailedKey], ShouldEqual, 0)
			So(results.Raw[SkipFailKey], ShouldEqual, 0)
			So(results.Raw[ProbesKey], ShouldEqual, 110647)
			So(results.Raw[SamplesKey], ShouldEqual, 125836)
			So(results.Raw[MinKey], ShouldEqual, 300)
			So(results.Raw[Percentile50Key], ShouldEqual, 2700)
			So(results.Raw[Percentile90Key], ShouldEqual, 6600)
			So(results.Raw[Percentile95Key], ShouldEqual, 35000)
			So(results.Raw[Percentile99Key], ShouldEqual, 352000)
			So(results.Raw[MaxKey], ShouldEqual, 1100000)
			So(results.Raw[QPSKey], ShouldEqual, 3999)
			So(results.Raw[IssuedRequestsKey], ShouldEqual, 3999)
		})
	})
	Convey("Attempting to read correct and readable file for latencies with output from SPECjbb run remotely", t, func() {
		path, err := filepath.Abs("remote_output")
		So(err, ShouldBeNil)
		Convey("should return correct results and no error", func() {
			results, err := FileWithLatencies(path)
			So(results.Raw, ShouldHaveLength, 14)
			So(err, ShouldBeNil)
			So(results.Raw[SuccessKey], ShouldEqual, 14355)
			So(results.Raw[PartialKey], ShouldEqual, 0)
			So(results.Raw[FailedKey], ShouldEqual, 0)
			So(results.Raw[SkipFailKey], ShouldEqual, 0)
			So(results.Raw[ProbesKey], ShouldEqual, 13832)
			So(results.Raw[SamplesKey], ShouldEqual, 15916)
			So(results.Raw[MinKey], ShouldEqual, 1000)
			So(results.Raw[Percentile50Key], ShouldEqual, 4000)
			So(results.Raw[Percentile90Key], ShouldEqual, 9530)
			So(results.Raw[Percentile95Key], ShouldEqual, 40150)
			So(results.Raw[Percentile99Key], ShouldEqual, 94000)
			So(results.Raw[MaxKey], ShouldEqual, 120000)
			So(results.Raw[QPSKey], ShouldEqual, 500)
			So(results.Raw[IssuedRequestsKey], ShouldEqual, 500)
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
	Convey("Opening readable and correct file for raw file name from remote output", t, func() {
		path, err := filepath.Abs("remote_output")
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
