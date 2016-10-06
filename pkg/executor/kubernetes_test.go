package executor

import (
	"testing"

	"k8s.io/kubernetes/pkg/kubelet/qos"

	. "github.com/smartystreets/goconvey/convey"
)

func TestKubernetes(t *testing.T) {
	Convey("Kubernetes config is sane by default", t, func() {
		config := DefaultKubernetesConfig()
		So(config.Privileged, ShouldEqual, false)
		So(config.HostNetwork, ShouldEqual, false)
		So(config.ContainerImage, ShouldEqual, defaultContainerImage)
	})

	Convey("After create new pod object", t, func() {
		config := DefaultKubernetesConfig()

		Convey("with default unspecified resources, expect BestEffort", func() {
			podExecutor := &kubernetes{config, nil}
			pod, err := podExecutor.newPod("be")
			So(err, ShouldBeNil)
			So(qos.GetPodQOS(pod), ShouldEqual, qos.BestEffort)
		})

		Convey("with CPU/Memory limit and requests euqal, expect Guaranteed", func() {
			config.CPURequest = 100
			config.CPULimit = 100
			config.MemoryRequest = 1000
			config.MemoryLimit = 1000
			podExecutor := &kubernetes{config, nil}
			pod, err := podExecutor.newPod("hp")
			So(err, ShouldBeNil)
			So(qos.GetPodQOS(pod), ShouldEqual, qos.Guaranteed)
		})

		Convey("with CPU/Memory limit and requests but not equal, expect Burstable", func() {
			config.CPURequest = 10
			config.CPULimit = 100
			config.MemoryRequest = 10
			config.MemoryLimit = 1000
			podExecutor := &kubernetes{config, nil}
			pod, err := podExecutor.newPod("burstable")
			So(err, ShouldBeNil)
			So(qos.GetPodQOS(pod), ShouldEqual, qos.Burstable)
		})

		Convey("with no CPU limit and request, expect Burstable", func() {
			config.CPURequest = 1
			config.CPULimit = 0
			podExecutor := &kubernetes{config, nil}
			pod, err := podExecutor.newPod("burst")
			So(err, ShouldBeNil)
			So(qos.GetPodQOS(pod), ShouldEqual, qos.Burstable)
		})

	})
}
