package executor

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/gophercloud/gophercloud"
)

func TestDefaultOpenstackConfig(t *testing.T) {

	Convey("OpenStack config is sane by default", t, func() {

		auth := gophercloud.AuthOptions{Username: "cirros"}
		config := DefaultOpenstackConfig(auth)

		So(config.Auth.Username, ShouldEqual, auth.Username)
		So(config.Flavor.Disk, ShouldEqual, 10)
		So(config.Flavor.RAM, ShouldEqual, 1024)
		So(config.Flavor.VCPUs, ShouldEqual, 1)
		So(config.Image, ShouldEqual, "cirros")
		So(config.User, ShouldEqual, "cirros")
		So(config.SSHKeyPath, ShouldEqual, "~/.ssh/id_rsa")
	})
}
