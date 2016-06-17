package scripts

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func runScript(parameter string) (string, error) {
	cmd := exec.Command("sh", "-c", os.Getenv("GOPATH")+
		"/src/github.com/intelsdi-x/swan/scripts/sensitivity-experiment.sh "+parameter)
	fmt.Println(cmd.Args)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func TestSensitivityExperimentScript(t *testing.T) {
	Convey("After running a script with help option", t, func() {
		output, err := runScript("-h")
		Convey("There should be no error and output should not be nil", func() {
			So(output, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
	})
	Convey("After running a script with non-existing experiment binary", t, func() {
		output, err := runScript("-p /nonExistingBinary")
		Convey("There should be an error and output should be empty", func() {
			So(output, ShouldEqual, "")
			So(fmt.Sprintf("%s", err), ShouldEqual, "exit status 127")
		})
	})

}
