package kubernetes_helpers

import (
	"k8s.io/client-go/1.5/pkg/api/resource"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/kubelet/qos"
)

// NOTE: functions defined here are extends client-go 1.5 functionality by
// functions 'IsPodReady' and 'GetPodQOS' that are needed by Swan thus basing on
// kubernetes 'master' branch were reimplemented to meet original functionality.

// IsPorReady returns true if Pod has condition ready fulfilled
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
	resNum = int(2)
	zeroQ  = resource.MustParse("0")
)

func isSupportedRes(res *v1.ResourceName) bool {
	if *res == v1.ResourceCPU || *res == v1.ResourceMemory {
		return true
	}
	return false
}

func processRequests(container *v1.Container, requests v1.ResourceList) {
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

func processLimits(container *v1.Container, limits v1.ResourceList) int {
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

func isGuaranteed(requests, limits v1.ResourceList, isQos bool) bool {
	if isQos {
		for rName, rVal := range requests {
			limit, ok := limits[rName]
			if !ok || limit.Cmp(rVal) != 0 {
				isQos = false
				break
			}
		}
	}
	if isQos && len(requests) == len(limits) {
		return true
	}
	return false
}

// GetPodQOS returns pod's QOS class.
// BestEffort - if none of it's containers have specified request or limits (CPU and Mem).
// Guaranteed - if across all containers request equals limits (CPU and Mem)
// Burstable - if across all containers requests and limits differs (CPU and Mem)
func GetPodQOS(pod *v1.Pod) qos.QOSClass {
	requests := v1.ResourceList{}
	limits := v1.ResourceList{}
	isQos := true
	for _, container := range pod.Spec.Containers {
		processRequests(&container, requests)

		lFound := processLimits(&container, limits)
		if lFound != resNum {
			isQos = false
		}
	}

	if len(requests) == 0 && len(limits) == 0 {
		return qos.BestEffort
	}

	if isGuaranteed(requests, limits, isQos) {
		return qos.Guaranteed
	}

	return qos.Burstable
}
