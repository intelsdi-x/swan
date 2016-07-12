package main

import (
	"github.com/intelsdi-x/swan/pkg/isolation/topo"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
)

// sharedCacheThreads returns threads from one socket that share a last-level
// cache. To avoid placing workloads on both hyperthreads for any physical
// core, only one thread from each is included in the result.
func sharedCacheThreads() topo.ThreadSet {
	allThreads, err := topo.Discover()
	errutil.Check(err)

	// Retain only threads for one socket.
	socket, err := allThreads.Sockets(1)
	errutil.Check(err)

	// Retain only one thread per physical core.
	// NB: The following filter prediccate closes over this int set.
	temp := socket.AvailableCores()

	return socket.Filter(func(t topo.Thread) bool {
		retain := temp.Contains(t.Core())
		temp.Remove(t.Core())
		return retain
	})
}

func getSiblingThreadsOfThread(reservedThread topo.Thread) topo.ThreadSet {
	requestedCore := reservedThread.Core()

	allThreads, err := topo.Discover()
	errutil.Check(err)

	threadsFromCore, err := allThreads.FromCores(requestedCore)
	errutil.Check(err)

	return threadsFromCore.Filter(func(t topo.Thread) bool {
		return t != reservedThread
	})
}

func getSiblingThreadsOfThreadSet(threads topo.ThreadSet) (results topo.ThreadSet) {
	for _, thread := range threads {
		siblings := getSiblingThreadsOfThread(thread)
		for _, sibling := range siblings {
			results = append(results, sibling)
		}
	}
	return results
}
