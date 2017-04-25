// Copyright (c) 2017 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package executor

import (
	"testing"

	"k8s.io/client-go/1.5/pkg/kubelet/qos"

	k8sports "github.com/intelsdi-x/swan/pkg/k8sports"
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
		config.ContainerImage = "centos_swan_image"

		Convey("with default unspecified resources, expect BestEffort", func() {
			podExecutor := &k8s{config, nil}
			pod, err := podExecutor.newPod("be")
			So(err, ShouldBeNil)
			So(k8sports.GetPodQOS(pod), ShouldEqual, qos.BestEffort)
		})

		Convey("with CPU/Memory limit and requests euqal, expect Guaranteed", func() {
			config.CPURequest = 100
			config.CPULimit = 100
			config.MemoryRequest = 1000
			config.MemoryLimit = 1000
			podExecutor := &k8s{config, nil}
			pod, err := podExecutor.newPod("hp")
			So(err, ShouldBeNil)
			So(k8sports.GetPodQOS(pod), ShouldEqual, qos.Guaranteed)
		})

		Convey("with CPU/Memory limit and requests but not equal, expect Burstable", func() {
			config.CPURequest = 10
			config.CPULimit = 100
			config.MemoryRequest = 10
			config.MemoryLimit = 1000
			podExecutor := &k8s{config, nil}
			pod, err := podExecutor.newPod("burstable")
			So(err, ShouldBeNil)
			So(k8sports.GetPodQOS(pod), ShouldEqual, qos.Burstable)
		})

		Convey("with no CPU limit and request, expect Burstable", func() {
			config.CPURequest = 1
			config.CPULimit = 0
			podExecutor := &k8s{config, nil}
			pod, err := podExecutor.newPod("burst")
			So(err, ShouldBeNil)
			So(k8sports.GetPodQOS(pod), ShouldEqual, qos.Burstable)
		})

	})

	Convey("Kubernetes pod executor pod names", t, func() {

		Convey("have desired name", func() {
			podExecutor := &k8s{KubernetesConfig{PodName: "foo"}, nil}
			name := podExecutor.generatePodName()
			So(name, ShouldEqual, "foo")

		})

		Convey("have desired prefix", func() {
			podExecutor := &k8s{KubernetesConfig{PodNamePrefix: "foo"}, nil}
			name := podExecutor.generatePodName()
			So(name, ShouldStartWith, "foo-")

		})

		Convey("with default config", func() {

			podExecutor := &k8s{DefaultKubernetesConfig(), nil}

			Convey("have default prefix", func() {
				name := podExecutor.generatePodName()
				So(name, ShouldStartWith, "swan-")
			})

			Convey("are unique", func() {
				names := make(map[string]struct{})
				N := 1000
				for i := 0; i < N; i++ {
					name := podExecutor.generatePodName()
					names[name] = struct{}{}
				}
				// Print(len(names))
				Print(names)
				So(names, ShouldHaveLength, N)
			})
		})
	})

}
