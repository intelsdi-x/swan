package executor

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
        "strings"
)

func TestRemote(t *testing.T){
	clientConfig := NewClientConfig("root", "/root/.ssh/id_rsa")
	sshConfig := NewSSHConfig(clientConfig, "localhost", 22)
	remoteShell := NewRemote(*sshConfig)
	convey.Convey("Using Remote Shell with empty command", t, func() {
		convey.Convey("Remote Shell", func() {
			remoteTask, _:= remoteShell.Execute("")
			remoteTask.Stop()
			_, taskStatus := remoteTask.Status()
			convey.Convey("Empty command gives empty output", func() {
					convey.So(taskStatus.stdout, convey.ShouldEqual, "")
			})
		})
	})
	convey.Convey("Using Remote Shell with whoami command", t, func() {
		convey.Convey("Remote Shell", func() {
			remoteTask, _ := remoteShell.Execute("whoami")
			remoteTask.Stop()
			_, taskStatus := remoteTask.Status()
			convey.Convey("whoami command with root user in clientconfig should give root as output", func() {
				convey.So(strings.TrimSpace(taskStatus.stdout), convey.ShouldEqual, "root")
			})
		})
	})
}
