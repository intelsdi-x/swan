// Copyright (c) 2017 Intel Corporation
// Copyright 2015 The Kubernetes Authors.
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
	"k8s.io/client-go/1.5/pkg/api/resource"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/kubelet/qos"
)

// NOTE: functions defined here are extends client-go 1.5 functionality by
// functions 'IsPodReady' and 'GetPodQOS' that are needed by Swan thus basing on
// kubernetes 'master' branch were reimplemented to meet original functionality.

// IsPodReady returns true if Pod has condition ready fulfilled
func IsPodReady(pod *v1.Pod) bool {
	for i := range pod.Status.Conditions {
		if pod.Status.Conditions[i].Type == v1.PodReady {
			if pod.Status.Conditions[i].Status == v1.ConditionTrue {
				return true
			}
		}
	}
	return false
}

var (
	// Resource CPU and Memory
	supportedResourceNumber = int(2)
	zeroQ                   = resource.MustParse("0")
)

// Supported Resources are: CPU and Memory.
func isSupportedRes(res *v1.ResourceName) bool {
	if *res == v1.ResourceCPU || *res == v1.ResourceMemory {
		return true
	}
	return false
}

// countRequests iterates over given container and counts the
// values for set Requests. The result is kept in 'request' ResourceList.
func countRequests(container *v1.Container, requests v1.ResourceList) {
	for reqName, reqQuantity := range container.Resources.Requests {
		if isSupportedRes(&reqName) {
			if reqQuantity.Cmp(zeroQ) == 1 {
				val := reqQuantity.Copy()
				_, ok := requests[reqName]
				if ok {
					val.Add(requests[reqName])
				}
				requests[reqName] = *val
			}
		}
	}
}

// countLimits iterates over given container and counts the
// values for set Limits. The result is kept in 'limits' ResourceList.
// It returns the number of found supported Limits.
func countLimits(container *v1.Container, limits v1.ResourceList) int {
	limitsFound := make(map[v1.ResourceName]int)
	for lName, lQuantity := range container.Resources.Limits {
		if isSupportedRes(&lName) {
			if lQuantity.Cmp(zeroQ) == 1 {
				limitsFound[lName] = 0
				val := lQuantity.Copy()
				_, ok := limits[lName]
				if ok {
					val.Add(limits[lName])
				}
				limits[lName] = *val
			}
		}
	}
	return len(limitsFound)
}

// Class is Guaranteed only if Requests equals Limits globally
// meaning that the same resources must be set and the values
// must be equal.
func isGuaranteed(requests, limits v1.ResourceList) bool {
	guaranteed := true
	for rName, rVal := range requests {
		limit, ok := limits[rName]
		if !ok || limit.Cmp(rVal) != 0 {
			guaranteed = false
			break
		}
	}
	if guaranteed && len(requests) == len(limits) {
		return true
	}
	return false
}

// GetPodQOS returns pod's QOS class.
// BestEffort - if none of it's containers have specified request or limits (CPU and Mem).
// Guaranteed - if across all containers request equals limits (CPU and Mem)
// Burstable - if across all containers requests and limits differs (CPU and Mem)
func GetPodQOS(pod *v1.Pod) qos.QOSClass {
	// Pod's Requests
	requests := v1.ResourceList{}
	// Pod's Limits
	limits := v1.ResourceList{}

	guaranteed := true
	for _, container := range pod.Spec.Containers {
		countRequests(&container, requests)

		limitsFound := countLimits(&container, limits)
		if limitsFound != supportedResourceNumber {
			// For Guaranteed all supported Resources need
			// to have limits set.
			guaranteed = false
		}
	}

	if len(requests) == 0 && len(limits) == 0 {
		return qos.BestEffort
	}

	// Check if 'Guaranteed' is fulfilled.
	if guaranteed && isGuaranteed(requests, limits) {
		return qos.Guaranteed
	}

	return qos.Burstable
}
