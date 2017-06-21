package topo

import (
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
)

// SharedCacheThreads returns threads from one socket that share a last-level
// cache. To avoid placing workloads on both hyperthreads for any physical
// core, only one thread from each is included in the result.
func SharedCacheThreads() ThreadSet {
	allThreads, err := Discover()
	errutil.Check(err)

	// Retain only threads for one socket.
	socket, err := allThreads.Sockets(1)
	errutil.Check(err)

	// Retain only one thread per physical core.
	// NB: The following filter prediccate closes over this int set.
	temp := socket.AvailableCores()

	return socket.Filter(func(t Thread) bool {
		retain := temp.Contains(t.Core())
		temp.Remove(t.Core())
		return retain
	})
}

// GetSiblingThreadsOfThread returns sibling HyperThread for supplied Thread.
func GetSiblingThreadsOfThread(reservedThread Thread) ThreadSet {
	requestedCore := reservedThread.Core()

	allThreads, err := Discover()
	errutil.Check(err)

	threadsFromCore, err := allThreads.FromCores(requestedCore)
	errutil.Check(err)

	return threadsFromCore.Remove(reservedThread)
}

// GetSiblingThreadsOfThreadSet returns sibling HyperThreads for supplied ThreadSet.
func GetSiblingThreadsOfThreadSet(threads ThreadSet) (results ThreadSet) {
	for _, thread := range threads {
		siblings := GetSiblingThreadsOfThread(thread)
		for _, sibling := range siblings {
			// Omit the reserved threads; if at least one pair from threads were
			// siblings of each other, they would both otherwise wrongly end up in
			// the result.
			if !threads.Contains(sibling) {
				results = append(results, sibling)
			}
		}
	}

	return results
}
