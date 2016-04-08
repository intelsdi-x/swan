package provisioning

import (
	. "github.com/smartystreets/goconvey/convey"
	"os/user"
	"strings"
	"testing"
)

func TestRemote(t *testing.T) {
	Convey("Creating a client configuration for the test user", t, func() {

		user, err := user.Current()
		So(err, ShouldBeNil)

		clientConfig := NewClientConfig(user.Username, user.HomeDir+"/.ssh/id_rsa")
		sshConfig := NewsshConfig(clientConfig, "localhost", 22)
		remoteShell := NewRemote(*sshConfig)

		Convey("Using Remote Shell with empty command", func() {
			Convey("Remote Shell", func() {
				remoteTask, _ := remoteShell.Run("")
				remoteTask.Status()
				Convey("Empty command gives empty output", func() {
					So(remoteTask.Status().stdout, ShouldEqual, "")
				})
			})
		})
		Convey("Using Remote Shell with whoami command", func() {
			Convey("Remote Shell", func() {
				remoteTask, err := remoteShell.Run("whoami")
				So(err, ShouldBeNil)

				remoteTask.Stop()
				Convey("whoami command with root user in clientconfig should give root as output", func() {
					So(strings.TrimSpace(remoteTask.Status().stdout), ShouldEqual, user.Username)
				})
			})
		})
	})
}
