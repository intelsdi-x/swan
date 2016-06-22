package main

import (
	"github.com/intelsdi-x/swan/pkg/isolation/topo"
)

// sharedCacheThreads returns threads from one socket that share a last-level
// cache. To avoid placing workloads on both hyperthreads for any physical
// core, only one thread from each is included in the result.
func sharedCacheThreads() topo.ThreadSet {
	allThreads, err := topo.Discover()
	check(err)

	// Retain only threads for one socket.
	socket, err := allThreads.Sockets(1)
	check(err)

	// Retain only one thread per physical core.
	// NB: The following filter prediccate closes over this int set.
	temp := socket.AvailableCores()

	return socket.Filter(func(t topo.Thread) bool {
		retain := temp.Contains(t.Core())
		temp.Remove(t.Core())
		return retain
	})
}
