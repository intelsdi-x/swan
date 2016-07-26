package testhelpers

import (
  "math/rand"
)

// RandomPorts returns count random ports between start and end.
func RandomPorts(start int, end int, count int) []int {
	ports := map[int]struct{}{}
	for len(ports) < count {
		port := rand.Intn(end-start) + start
		ports[port] = struct{}{}
	}

	out := []int{}
	for port, _ := range ports {
		out = append(out, port)
	}

	return out
}
