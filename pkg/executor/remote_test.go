// +build integration

package executor

import (
	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os"
	"os/user"
	"regexp"
	"testing"
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
			if _, err := os.Stat(user.HomeDir + "/.ssh/id_rsa"); os.IsNotExist(err) {
				t.Skip("skipping test: ssh keys not found in " + user.HomeDir + "/.ssh/id_rsa")
			}

			// Check if localhost is self-authorized.
			hostname, err := os.Hostname()
			if err != nil {
				t.Skip("skipping test: cannot figure out if localhost is self-authorized")
			}

			authorizedHostsFile, err := os.Open(user.HomeDir + "/.ssh/authorized_keys")
			if err != nil {
				t.Skip("skipping test: cannot figure out if localhost is self-authorized", err)
			}
			authorizedHosts, err := ioutil.ReadAll(authorizedHostsFile)
			if err != nil {
				t.Skip("skipping test: cannot figure out if localhost is self-authorized", err)
			}

			re := regexp.MustCompile(hostname)
			match := re.Find(authorizedHosts)

			if match == nil {
				t.Skip("skipping test: localhost (" + hostname + ") is not self-authorized")
			}

			clientConfig, err := NewClientConfig(user.Username, user.HomeDir+"/.ssh/id_rsa")

			Convey("There should be no error", func() {
				So(err, ShouldBeNil)
			})

			Convey("And while using Remote Shell and connection to localhost", func() {
				sshConfig := NewSSHConfig(clientConfig, "localhost", 22)

				Convey("The generic Executor test should pass", func() {
					testExecutor(t, NewRemote(*sshConfig))
				})
			})
		})

	})
}
