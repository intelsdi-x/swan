package executor

import (
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
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
			if _, err := os.Stat(user.HomeDir + "/.ssh/id_rsa"); os.IsNotExist(err) {
				t.Skip("skipping test: ssh keys not found in " + user.HomeDir + "/.ssh/id_rsa")
			}
			clientConfig, err := NewClientConfig(user.Username, user.HomeDir+"/.ssh/id_rsa")
			So(err, ShouldBeNil)
			sshConfig := NewSSHConfig(clientConfig, "localhost", 22)
			remoteShell := NewRemote(*sshConfig)
			Convey("with not existing command", func() {
				remoteTask, _ := remoteShell.Execute("notexistingcommand")
				remoteTask.Stop()
				_, taskStatus := remoteTask.Status()
				Convey("should give 127 error ExitCode", func() {
					So(*taskStatus, ShouldEqual, 127)
				})
			})
			Convey("with whoami command ", func() {
				remoteTask, err := remoteShell.Execute("whoami")
				So(err, ShouldBeNil)
				remoteTask.Stop()
				stdoutReader, err := remoteTask.Stdout()
				So(err, ShouldBeNil)
				data, err := ioutil.ReadAll(stdoutReader)
				So(err, ShouldBeNil)
				Convey("with root user in clientconfig should give root as output", func() {
					So(strings.TrimSpace(string(data[:])), ShouldEqual, user.Username)
				})
				fileName, err := remoteTask.GetStdoutDir()
				So(err, ShouldBeNil)
				fileInf, err := os.Stat(fileName)
				Convey("before cleaning file should exist", func() {
					So("/tmp/"+fileInf.Name(), ShouldEqual, fileName)
				})
				remoteTask.Clean()
				_, err = os.Stat(fileName)
				Convey("after cleaning file should not exist", func() {
					So(err, ShouldNotBeNil)
				})

			})
		})

	})
}
