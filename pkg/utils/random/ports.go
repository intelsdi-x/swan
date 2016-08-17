package random

import (
	"math/rand"
)

// Ports returns count random ports between start and end.
func Ports(start int, end int, count int) []int {
	ports := map[int]struct{}{}
	for len(ports) < count {
		port := rand.Intn(end-start) + start
		ports[port] = struct{}{}
	}

	out := []int{}
	for port := range ports {
		out = append(out, port)
	}

	return out
}
