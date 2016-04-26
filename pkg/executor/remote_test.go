package executor

import (
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
	"testing"
)

func getRemoteStdoutPath(task *remoteTask) (string, error) {
	if _, err := os.Stat(task.stdoutFile.Name()); err != nil {
		return "", err
	}
	return task.stdoutFile.Name(), nil
}

func getRemoteStderrPath(task *remoteTask) (string, error) {
	if _, err := os.Stat(task.stderrFile.Name()); err != nil {
		return "", err
	}
	return task.stderrFile.Name(), nil
}

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
				task, _ := remoteShell.Execute("notexistingcommand")
				task.Stop()
				_, taskStatus := task.Status()
				Convey("should give 127 error ExitCode", func() {
					So(*taskStatus, ShouldEqual, 127)
				})
			})
			Convey("with whoami command ", func() {
				task, err := remoteShell.Execute("whoami")
				So(err, ShouldBeNil)
				task.Stop()
				stdoutReader := task.Stdout()
				data, err := ioutil.ReadAll(stdoutReader)
				So(err, ShouldBeNil)
				Convey("with root user in clientconfig should give root as output", func() {
					So(strings.TrimSpace(string(data[:])), ShouldEqual, user.Username)
				})
				pwd, err := os.Getwd()
				So(err, ShouldBeNil)
				fileName, err := getRemoteStdoutPath(task.(*remoteTask))
				So(err, ShouldBeNil)
				fileInf, err := os.Stat(fileName)
				Convey("before cleaning file should exist", func() {
					So(pwd+"/"+fileInf.Name(), ShouldEqual, fileName)
				})
				//task.Clean()
				//_, err = os.Stat(fileName)
				//Convey("after cleaning file should not exist", func() {
				//	So(err, ShouldNotBeNil)
				//})

			})
		})

	})
}
