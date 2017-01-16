package executor

import (
	"fmt"
	"os"
	"path"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	//expectedFileMode is a string equivalent of 0644
	expectedFileMode = "-rw-r--r--"
	//expectedDirMode is a string equivalent of 0755
	expectedDirMode = "drwxr-xr-x"
)

func filesCleanup(stdout, stderr *os.File) {
	err := stderr.Close()
	So(err, ShouldBeNil)
	err = stdout.Close()
	So(err, ShouldBeNil)

	err = os.RemoveAll(path.Dir(stdout.Name()))
	So(err, ShouldBeNil)

}

func TestBinaryNameFromCommand(t *testing.T) {
	testData := map[string]string{
		"/ust/bin/bash":                                           "bash",
		"/ust/bin/bash -a dsf -b fsdf -c sdgfs -xx fwef":          "bash",
		"/ust/bin/bash --fancy-option=sfsdfsdf":                   "bash",
		"/ust/bin/bash --fancy-option2=http://123.123.45.34:3242": "bash",
		"/ust/bin/bash --fancy-option3=123.123.0.0/16":            "bash",
	}

	for command, expectedResult := range testData {
		Convey(fmt.Sprintf("I should get the binary name = %q from command = %q", expectedResult, command), t, func() {
			binaryName, err := getBinaryNameFromCommand(command)
			So(err, ShouldBeNil)
			So(binaryName, ShouldEqual, expectedResult)
		})
	}
}

func TestCreateExecutorOutputFiles(t *testing.T) {
	Convey("I should be able to create files and folders for experiment details", t, func() {
		outputDir, err := createOutputDirectory("command", "test")
		So(err, ShouldBeNil)
		stdout, stderr, err := createExecutorOutputFiles(outputDir)
		So(err, ShouldBeNil)
		So(stdout, ShouldNotBeNil)
		So(stderr, ShouldNotBeNil)

		defer filesCleanup(stdout, stderr)

		Convey("Which should have got valid modes.", func() {
			eStat, err := stderr.Stat()
			So(err, ShouldBeNil)
			So(eStat.Mode().String(), ShouldEqual, expectedFileMode)

			oStat, err := stdout.Stat()
			So(err, ShouldBeNil)
			So(oStat.Mode().String(), ShouldEqual, expectedFileMode)

			parentDir := path.Dir(stdout.Name())
			pDirStat, err := os.Stat(parentDir)
			So(err, ShouldBeNil)
			So(pDirStat.Mode().String(), ShouldEqual, expectedDirMode)
		})
	})
}
