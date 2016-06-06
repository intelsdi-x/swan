package executor_test

import (
	"os/user"
	"syscall"
	"testing"

	log "github.com/Sirupsen/logrus"
	. "github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/isolation"
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
			if err := CheckDefaultRemoteConfigRequirements(user); err != nil {
				t.Skip(err)
			}

			clientConfig, err := NewClientConfig(user.Username, user.HomeDir+"/.ssh/id_rsa")

			Convey("There should be no error", func() {
				So(err, ShouldBeNil)
			})

			Convey("And while using Remote Shell and connection to localhost", func() {
				sshConfig := NewSSHConfig(clientConfig, "localhost", 22)
				isolationPid, err := isolation.NewNamespace(syscall.CLONE_NEWPID)
				So(err, ShouldBeNil)

				Convey("The generic Executor test should pass", func() {
					testExecutor(t, NewRemote(*sshConfig, isolationPid))
				})
			})
		})

	})
}
