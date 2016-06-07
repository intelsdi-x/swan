package executor_test

import (
	"os/user"
	"testing"

	log "github.com/Sirupsen/logrus"
	. "github.com/intelsdi-x/swan/pkg/executor"
	. "github.com/smartystreets/goconvey/convey"
)

// This tests required following setup:
// - id_rsa ssh keys in user home directory. [command ssh-keygen]
// - no password ssh session. [command ssh-copy-id localhost]
func TestRemote(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("While getting the information abut the test user", t, func() {
		user, err := user.Current()

		Convey("There should be no error", func() {
			So(err, ShouldBeNil)
		})

		Convey("When creating ssh configuration with proper data", func() {
			sshConfig, err := NewSSHConfig("127.0.0.1", DefaultSSHPort, user)
			if err != nil {
				// Skip test if setup is not wel configured.
				t.Skip("Skipping test: " + err.Error())
			}

			Convey("There should be no error", func() {
				So(err, ShouldBeNil)
			})

			Convey("And while using Remote Shell, the generic Executor test should pass", func() {
				testExecutor(t, NewRemote(*sshConfig))
			})
		})
	})
}
