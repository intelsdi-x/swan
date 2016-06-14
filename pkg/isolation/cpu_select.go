package isolation

import (
	"errors"
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
)

// ShareLLCButNotL1L2 filter for selecting cpu.
const ShareLLCButNotL1L2 = 1 << iota

// Implements round-robin allocation of cores.
// TODO: Improve this, potentially keeping track of allocations
//       on the system.
var allocationSocket = 0
var allocationSocketMutex sync.Mutex

func nextSocket() int {
	allocationSocketMutex.Lock()
	defer allocationSocketMutex.Unlock()

	var info CPUInfo
	err := info.Discover()
	if err != nil {
		return allocationSocket
	}
	next := allocationSocket
	allocationSocket = (allocationSocket + 1) % info.Sockets
	return next
}

// CPUSelect returns a set of logical cpu ids that match the supplied
// criteria.
//
// For now, the only supported filter is to select cores that share LLC
// but do not share L1 or L2 cache.
//
// Returns an error if the request cannot be satisfied.
func CPUSelect(countRequested int, filters uint) (IntSet, error) {
	if countRequested == 0 {
		return nil, errors.New("Number of core requested is zero")
	}

	// info collect CPU topology.
	var info CPUInfo
	err := info.Discover()
	if err != nil {
		return nil, err
	}

	if filters == ShareLLCButNotL1L2 {
		for i := 0; i < info.Sockets; i++ {
			socket := nextSocket()
			cpus, err := searchSocket(info, countRequested, socket)
			if err == nil {
				if len(cpus) == countRequested {
					log.Debug("Answering CPUSelect query for %d cpus with %v on socket %d", countRequested, cpus, socket)
					return cpus, nil
				}
			}
		}

		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Unsatisfiable request")
	}

	return nil, fmt.Errorf("Unknown filter supplied (%d)", filters)
}

func searchSocket(info CPUInfo, countRequested, socket int) (IntSet, error) {
	result := NewIntSet()
	cores := info.SocketCores[socket]

	if countRequested > len(cores) {
		return nil, fmt.Errorf("Unsatisfiable request: need %d cores but only have %d", countRequested, len(cores))
	}

	for c := range cores {
		if len(result) == countRequested {
			break
		}
		cpus := info.CoreCpus[c]
		// NOTE: Go map iteration order is intentionally randomized
		//       to prevent depending on any implied order.
		for cpu := range cpus {
			// Add at most one logical cpu from each physical core
			result.Add(cpu)
			break
		}
	}
	return result, nil
}
