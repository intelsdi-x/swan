package scripts

import (
	"fmt"
	"github.com/intelsdi-x/swan/pkg/cassandra"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func runScript(parameter string) (string, error) {
	cmd := exec.Command("sh", "-c", os.Getenv("GOPATH")+
		"/src/github.com/intelsdi-x/swan/scripts/sensitivity-experiment.sh "+parameter)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func TestSensitivityExperimentScript(t *testing.T) {
	Convey("While connecting to Cassandra with proper parameters", t, func() {
		cassandraConfig, err := cassandra.CreateConfigWithSession("127.0.0.1", "snap")
		So(err, ShouldBeNil)
		session := cassandraConfig.CassandraSession()
		Convey("I should receive not empty session", func() {
			So(session, ShouldNotBeNil)
			So(err, ShouldBeNil)
			Convey("After running a script with help option", func() {
				output, err := runScript("-h")
				Convey("There should be no error and output should not be nil", func() {
					So(output, ShouldNotBeNil)
					So(err, ShouldBeNil)
				})
			})
			Convey("After running a script with non-existing experiment binary", func() {
				output, err := runScript("-p /nonExistingBinary")
				Convey("There should be an error and output should be empty", func() {
					So(output, ShouldEqual, "")
					So(fmt.Sprintf("%s", err), ShouldEqual, "exit status 127")
				})
			})
			Convey("After running a script with existing proper experiment", func() {
				output, err := runScript("-p integration_tests/pkg/scripts/fakeExperiment.sh")
				Convey("There should be no error and output should not be empty", func() {
					So(output, ShouldNotBeNil)
					So(err, ShouldBeNil)
				})
			})

		})
	})
}
