package provisioning

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
        "strings"
)

func TestRemote(t *testing.T){
	clientConfig := NewClientConfig("root", "/root/.ssh/id_rsa")
	sshConfig := NewsshConfig(clientConfig, "localhost", 22)
	remoteShell := NewRemote(*sshConfig)
	convey.Convey("Using Remote Shell with empty command", t, func() {
		convey.Convey("Remote Shell", func() {
			remoteTask, _:= remoteShell.Run("")
			remoteTask.Status()
			convey.Convey("Empty command gives empty output", func() {
					convey.So(remoteTask.Status().stdout, convey.ShouldEqual, "")
			})
		})
	})
	convey.Convey("Using Remote Shell with whoami command", t, func() {
		convey.Convey("Remote Shell", func() {
			remoteTask, _ := remoteShell.Run("whoami")
			remoteTask.Stop()
			convey.Convey("whoami command with root user in clientconfig should give root as output", func() {
				convey.So(strings.TrimSpace(remoteTask.Status().stdout), convey.ShouldEqual, "root")
			})
		})
	})
}
