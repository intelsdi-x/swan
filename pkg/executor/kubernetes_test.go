package executor

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestKubernetes(t *testing.T) {
	Convey("Kubernetes config is sane by default", t, func() {
		config := DefaultKubernetesConfig()
		So(config.Privileged, ShouldEqual, false)
		So(config.HostNetwork, ShouldEqual, false)
		So(config.ContainerImage, ShouldEqual, defaultContainerImage)
	})
}
