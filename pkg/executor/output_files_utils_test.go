package executor

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"os"
	"path"
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

func TestCreateExecutorOutputFiles(t *testing.T) {

	Convey("I should be able to create files and folders for experiment details", t, func() {
		stdout, stderr, err := createExecutorOutputFiles("command", "test")
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
