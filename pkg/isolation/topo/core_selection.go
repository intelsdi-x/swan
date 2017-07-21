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
