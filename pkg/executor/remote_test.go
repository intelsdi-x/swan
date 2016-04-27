package executor

import (
	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"os/user"
	"testing"
)

// This tests required following setup:
// - id_rsa ssh keys in user home directory.
// - ssh-copy-id localhost
func TestRemote(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("While getting the information abut the test user", t, func() {
		user, err := user.Current()

		Convey("There should be no error", func() {
			So(err, ShouldBeNil)
		})

		Convey("When creating client configuration with not existing path to key file", func() {
			clientConfig, err := NewClientConfig(user.Username, user.HomeDir+"/notexistingdirectoryname")

			Convey("should rise error", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("should result in nil client config", func() {
				So(clientConfig, ShouldBeNil)
			})
		})

		Convey("When creating client configuration with proper configuration", func() {
			if _, err := os.Stat(user.HomeDir + "/.ssh/id_rsa"); os.IsNotExist(err) {
				t.Skip("skipping test: ssh keys not found in " + user.HomeDir + "/.ssh/id_rsa")
			}

			clientConfig, err := NewClientConfig(user.Username, user.HomeDir+"/.ssh/id_rsa")

			Convey("There should be no error", func() {
				So(err, ShouldBeNil)
			})

			Convey("And while using Remote Shell and connection to localhost", func() {
				sshConfig := NewSSHConfig(clientConfig, "localhost", 22)

				Convey("The generic Executor test should pass", func() {
					TestExecutor(t, NewRemote(*sshConfig))
				})
			})
		})

	})
}
