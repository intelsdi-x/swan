package random

import (
	"math/rand"
	"time"
)

var source int64

// Ports returns count random ports between start and end.
func Ports(start int, end int, count int) []int {
	if source == 0 {
		source = time.Now().UnixNano()
	} else {
		source = source + 1
	}
	r := rand.New(rand.NewSource(source))
	ports := map[int]struct{}{}
	for len(ports) < count {
		port := r.Intn(end-start) + start
		ports[port] = struct{}{}
	}

	out := []int{}
	for port := range ports {
		out = append(out, port)
	}

	return out
}
