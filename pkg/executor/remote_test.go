package executor

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"os/user"
	"strings"
	"testing"
)

func TestRemote(t *testing.T) {
	SkipConvey("Creating a client configuration for the test user", t, func() {
		user, err := user.Current()
		So(err, ShouldBeNil)
		Convey("Using Remote Shell with not existing path to key file", func() {
			clientConfig, err := NewClientConfig(user.Username, user.HomeDir+"/notexistingdirectoryname")
			Convey("should rise error", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("should result in nil client config", func() {
				So(clientConfig, ShouldBeNil)
			})
		})
		Convey("Using Remote Shell with proper configuration", func() {
			if _, err := os.Stat(user.HomeDir+"/.ssh/id_rsa"); os.IsNotExist(err) {
				t.Skip("skipping test: ssh keys not found in "+user.HomeDir+"/.ssh/id_rsa")
			}
			clientConfig, err := NewClientConfig(user.Username, user.HomeDir+"/.ssh/id_rsa")
			So(err, ShouldBeNil)
			sshConfig := NewSSHConfig(clientConfig, "localhost", 22)
			remoteShell := NewRemote(*sshConfig)
			Convey("with empty command", func() {
				remoteTask, _ := remoteShell.Execute("")
				remoteTask.Stop()
				_, taskStatus := remoteTask.Status()
				Convey("gives empty output", func() {
					So(taskStatus.Stdout, ShouldEqual, "")
				})
			})
			Convey("with not existing command", func() {
				remoteTask, _ := remoteShell.Execute("notexistingcommand")
				remoteTask.Stop()
				_, taskStatus := remoteTask.Status()
				Convey("should give 127 error ExitCode", func() {
					So(taskStatus.ExitCode, ShouldEqual, 127)
				})
			})
			Convey("with whoami command ", func() {
				remoteTask, err := remoteShell.Execute("whoami")
				So(err, ShouldBeNil)
				remoteTask.Stop()
				_, taskStatus := remoteTask.Status()
				Convey("with root user in clientconfig should give root as output", func() {
					So(strings.TrimSpace(taskStatus.Stdout), ShouldEqual, user.Username)
				})
			})
		})

	})
}
