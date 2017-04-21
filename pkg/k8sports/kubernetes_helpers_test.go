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

package k8sports

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/client-go/1.5/pkg/api/resource"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/kubelet/qos"
)

func newRes(cpu, mem string) v1.ResourceList {
	res := v1.ResourceList{}
	if cpu != "" {
		res[v1.ResourceCPU] = resource.MustParse(cpu)
	}
	if mem != "" {
		res[v1.ResourceMemory] = resource.MustParse(mem)
	}
	return res
}

func addRes(name, val string, rlist v1.ResourceList) v1.ResourceList {
	rlist[v1.ResourceName(name)] = resource.MustParse(val)
	return rlist
}

func newContainer(name string, req, limits v1.ResourceList) v1.Container {
	return v1.Container{
		Name: name,
		Resources: v1.ResourceRequirements{
			Requests: req,
			Limits:   limits,
		},
	}
}

func newPod(name string, containers []v1.Container) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: v1.PodSpec{
			Containers: containers,
		},
	}
}

func TestGetPodQOS(t *testing.T) {

	// Checks against BestEffort
	Convey("A Pod with single Container without limits and requests set shall have BestEffort class", t, func() {
		class := GetPodQOS(newPod("BestEffort-Pod", []v1.Container{
			newContainer("BestEffort-Container", newRes("", ""), newRes("", "")),
		}))
		So(class, ShouldEqual, qos.BestEffort)
	})

	Convey("A Pod with four Containers without limits and requests set shall have BestEffort class", t, func() {
		class := GetPodQOS(newPod("BestEffort-Pod", []v1.Container{
			newContainer("BestEffort-Container1", newRes("", ""), newRes("", "")),
			newContainer("BestEffort-Container2", newRes("", ""), newRes("", "")),
			newContainer("BestEffort-Container3", newRes("", ""), newRes("", "")),
			newContainer("BestEffort-Container4", newRes("", ""), newRes("", "")),
		}))
		So(class, ShouldEqual, qos.BestEffort)
	})

	Convey("A Pod with single Container without limits/request for CPU and Memory but set for Storage set shall have BestEffort class", t, func() {
		class := GetPodQOS(newPod("BestEffort-Pod", []v1.Container{
			newContainer("BestEffort-Container", newRes("", ""), addRes("storage", "2Gi", newRes("", ""))),
		}))
		So(class, ShouldEqual, qos.BestEffort)
	})

	Convey("A Pod has four Containers. All containers have CPU and Memory unset. One container has additional resource. Pod shall have BestEffort class", t, func() {
		class := GetPodQOS(newPod("BestEffort-Pod", []v1.Container{
			newContainer("BestEffort-Container1", newRes("", ""), newRes("", "")),
			newContainer("BestEffort-Container2", newRes("", ""), newRes("", "")),
			newContainer("BestEffort-Container3", newRes("", ""), newRes("", "")),
			newContainer("BestEffort-Container4", newRes("", ""), addRes("storage", "2Gi", newRes("", ""))),
		}))
		So(class, ShouldEqual, qos.BestEffort)
	})

	// Checks against Guaranteed
	Convey("A Pod with single Container with equal limits and requests set shall have Guaranteed class", t, func() {
		class := GetPodQOS(newPod("Guaranteed-Pod", []v1.Container{
			newContainer("Guaranteed-Container", newRes("100m", "100Mi"), newRes("100m", "100Mi")),
		}))
		So(class, ShouldEqual, qos.Guaranteed)
	})

	Convey("A Pod with four Containers with equal limits and requests across all containers set shall have Guaranteed class", t, func() {
		class := GetPodQOS(newPod("Guaranteed-Pod", []v1.Container{
			newContainer("Guaranteed-Container1", newRes("100m", "100Mi"), newRes("100m", "100Mi")),
			newContainer("Guaranteed-Container2", newRes("100m", "100Mi"), newRes("100m", "100Mi")),
			newContainer("Guaranteed-Container3", newRes("100m", "100Mi"), newRes("100m", "100Mi")),
			newContainer("Guaranteed-Container4", newRes("100m", "100Mi"), newRes("100m", "100Mi")),
		}))
		So(class, ShouldEqual, qos.Guaranteed)
	})

	Convey("A Pod has CPU and Memory set for Guaranteed. Additional Storage request/limit shall not change the Qos class", t, func() {
		class := GetPodQOS(newPod("Guaranteed-Pod", []v1.Container{
			newContainer("Guaranteed-Container", newRes("100m", "100Mi"), addRes("storage", "2Gi", newRes("100m", "100Mi"))),
		}))
		So(class, ShouldEqual, qos.Guaranteed)
	})

	Convey("A Pod with four Containers. CPU and Memory Set for Guaranteed. Two containers have addtitional resources nvidia-gpu and storage. Class shall be Guaranteed", t, func() {
		class := GetPodQOS(newPod("Guaranteed-Pod", []v1.Container{
			newContainer("Guaranteed-Container1", newRes("100m", "100Mi"), newRes("100m", "100Mi")),
			newContainer("Guaranteed-Container2", newRes("100m", "100Mi"), addRes("storage", "1Gi", newRes("100m", "100Mi"))),
			newContainer("Guaranteed-Container3", newRes("100m", "100Mi"), newRes("100m", "100Mi")),
			newContainer("Guaranteed-Container4", newRes("100m", "100Mi"), addRes("nvidia-gpu", "2", newRes("100m", "100Mi"))),
		}))
		So(class, ShouldEqual, qos.Guaranteed)
	})

	Convey("A Pod with four Containers. CPU and Memory Set for Guaranteed. Class shall be Guaranteed", t, func() {
		class := GetPodQOS(newPod("Guaranteed-Pod", []v1.Container{
			newContainer("Guaranteed-Container1", newRes("100m", "100Mi"), newRes("100m", "100Mi")),
			newContainer("Guaranteed-Container2", newRes("100m", "100Mi"), newRes("100m", "100Mi")),
			newContainer("Guaranteed-Container3", newRes("100m", "100Mi"), newRes("100m", "100Mi")),
			newContainer("Guaranteed-Container4", newRes("100m", "100Mi"), newRes("100m", "100Mi")),
		}))
		So(class, ShouldEqual, qos.Guaranteed)
	})

	// User shall not be allowed to pass request > limits. However this show how algorithm works:
	// by summing up values in all containers
	Convey("A Pod has CPU and Memory set for Burstable but it will be classified as Guaranteed", t, func() {
		class := GetPodQOS(newPod("Guaranteed-Pod", []v1.Container{
			newContainer("Guaranteed-Container1", newRes("45m", "55Mi"), newRes("55m", "45Mi")),
			newContainer("Guaranteed-Container2", newRes("55m", "45Mi"), newRes("45m", "55Mi")),
		}))
		So(class, ShouldEqual, qos.Guaranteed)
	})

	// Check against Burstable
	Convey("A Pod with single Container with request/limits set to meet Burstable class should have Burstable class", t, func() {
		class := GetPodQOS(newPod("Burstable-Pod", []v1.Container{
			newContainer("Burstable-Container", newRes("", "100Mi"), newRes("", "")),
		}))
		So(class, ShouldEqual, qos.Burstable)

		class = GetPodQOS(newPod("Burstable-Pod", []v1.Container{
			newContainer("Burstable-Container", newRes("100m", ""), newRes("", "")),
		}))
		So(class, ShouldEqual, qos.Burstable)

		class = GetPodQOS(newPod("Burstable-Pod", []v1.Container{
			newContainer("Burstable-Container", newRes("", ""), newRes("100m", "")),
		}))
		So(class, ShouldEqual, qos.Burstable)

		class = GetPodQOS(newPod("Burstable-Pod", []v1.Container{
			newContainer("Burstable-Container", newRes("", ""), newRes("", "100Mi")),
		}))
		So(class, ShouldEqual, qos.Burstable)

		class = GetPodQOS(newPod("Burstable-Pod", []v1.Container{
			newContainer("Burstable-Container", newRes("100m", "100Mi"), newRes("", "")),
		}))
		So(class, ShouldEqual, qos.Burstable)

		class = GetPodQOS(newPod("Burstable-Pod", []v1.Container{
			newContainer("Burstable-Container", newRes("100m", "100Mi"), newRes("100m", "")),
		}))
		So(class, ShouldEqual, qos.Burstable)

		class = GetPodQOS(newPod("Burstable-Pod", []v1.Container{
			newContainer("Burstable-Container", newRes("100m", "100Mi"), newRes("", "100Mi")),
		}))
		So(class, ShouldEqual, qos.Burstable)

		class = GetPodQOS(newPod("Burstable-Pod", []v1.Container{
			newContainer("Burstable-Container", newRes("100m", ""), newRes("100m", "100Mi")),
		}))
		So(class, ShouldEqual, qos.Burstable)

		class = GetPodQOS(newPod("Burstable-Pod", []v1.Container{
			newContainer("Burstable-Container", newRes("100m", ""), addRes("storage", "2Gi", newRes("100m", "100Mi"))),
		}))
		So(class, ShouldEqual, qos.Burstable)

		class = GetPodQOS(newPod("Burstable-Pod", []v1.Container{
			newContainer("Burstable-Container", newRes("100m", "99Mi"), newRes("100m", "100Mi")),
		}))
		So(class, ShouldEqual, qos.Burstable)

		class = GetPodQOS(newPod("Burstable-Pod", []v1.Container{
			newContainer("Burstable-Container", newRes("100m", "99Mi"), addRes("storage", "2Gi", newRes("100m", "100Mi"))),
		}))
		So(class, ShouldEqual, qos.Burstable)

		class = GetPodQOS(newPod("Burstable-Pod", []v1.Container{
			newContainer("Burstable-Container", newRes("99m", "100Mi"), newRes("100m", "100Mi")),
		}))
		So(class, ShouldEqual, qos.Burstable)

		class = GetPodQOS(newPod("Burstable-Pod", []v1.Container{
			newContainer("Burstable-Container", newRes("100m", "100Mi"), newRes("99m", "100Mi")),
		}))
		So(class, ShouldEqual, qos.Burstable)

		class = GetPodQOS(newPod("Burstable-Pod", []v1.Container{
			newContainer("Burstable-Container", newRes("100m", "100Mi"), newRes("100m", "99Mi")),
		}))
		So(class, ShouldEqual, qos.Burstable)

		class = GetPodQOS(newPod("Burstable-Pod", []v1.Container{
			newContainer("Burstable-Container1", newRes("100m", "100Mi"), newRes("100m", "100Mi")),
			newContainer("Burstable-Container2", newRes("100m", "99Mi"), newRes("100m", "100Mi")),
		}))
		So(class, ShouldEqual, qos.Burstable)

		class = GetPodQOS(newPod("Burstable-Pod", []v1.Container{
			newContainer("Burstable-Container1", newRes("100m", "100Mi"), newRes("100m", "100Mi")),
			newContainer("Burstable-Container2", newRes("100m", "99Mi"), addRes("storage", "2Gi", newRes("100m", "100Mi"))),
		}))
		So(class, ShouldEqual, qos.Burstable)

		class = GetPodQOS(newPod("Burstable-Pod", []v1.Container{
			newContainer("Burstable-Container1", newRes("", ""), newRes("", "")),
			newContainer("Burstable-Container2", newRes("", "99Mi"), addRes("storage", "2Gi", newRes("", ""))),
		}))
		So(class, ShouldEqual, qos.Burstable)

	})

}
