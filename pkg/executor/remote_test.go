package executor

import (
	"github.com/smartystreets/goconvey/convey"
	"os/user"
	"strings"
	"testing"
)

func TestRemote(t *testing.T) {
	convey.SkipConvey("Creating a client configuration for the test user", t, func() {
		user, err := user.Current()
		convey.So(err, convey.ShouldBeNil)
		convey.Convey("Using Remote Shell with proper configuration", func() {
			clientConfig, err := NewClientConfig(user.Username, user.HomeDir+"/.ssh/id_rsa")
			convey.So(err, convey.ShouldBeNil)
			sshConfig := NewSSHConfig(clientConfig, "localhost", 22)
			remoteShell := NewRemote(*sshConfig)
			convey.Convey("with empty command", func() {
				remoteTask, _ := remoteShell.Execute("")
				remoteTask.Stop()
				_, taskStatus := remoteTask.Status()
				convey.Convey("gives empty output", func() {
					convey.So(taskStatus.Stdout, convey.ShouldEqual, "")
				})
			})
			convey.Convey("with not existing command", func() {
				remoteTask, _ := remoteShell.Execute("notexistingcommand")
				remoteTask.Stop()
				_, taskStatus := remoteTask.Status()
				convey.Convey("should give 127 error ExitCode", func() {
					convey.So(taskStatus.ExitCode, convey.ShouldEqual, 127)
				})
			})
			convey.Convey("with whoami command ", func() {
				remoteTask, err := remoteShell.Execute("whoami")
				convey.So(err, convey.ShouldBeNil)
				remoteTask.Stop()
				_, taskStatus := remoteTask.Status()
				convey.Convey("with root user in clientconfig should give root as output", func() {
					convey.So(strings.TrimSpace(taskStatus.Stdout), convey.ShouldEqual, user.Username)
				})
			})
		})
		convey.Convey("Using Remote Shell with not existing path to key file", func() {
			clientConfig, err := NewClientConfig(user.Username, user.HomeDir+"/notexistingdirectoryname")
			convey.Convey("should rise error", func() {
				convey.So(err, convey.ShouldNotBeNil)
			})
			convey.Convey("should result in nil client config", func() {
				convey.So(clientConfig, convey.ShouldBeNil)
			})
		})

	})
}
