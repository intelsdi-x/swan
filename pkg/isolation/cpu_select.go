package isolation

import (
	"errors"
)

// const filter for selecting cpu
const	ShareLLCButNotL1L2     uint = 1


// CPUIds represents a set of HW thread Ids.
type CPUIds map[int]int

// ThreadSet defines data needed for selected CPU.
type ThreadSet struct {
	requestedThreads CPUIds
	requestID	 string
}


// NewThreadSet instance creation.
func NewThreadSet(cpuids CPUIds, reqid string) *ThreadSet {
	return &ThreadSet{requestedThreads: cpuids, requestID: reqid}
}

// CPUSelect with the desired criterion.
func (threadset *ThreadSet) CPUSelect(count int, filters uint) error {
     // cpuDiscovered collect CPU topology.
     var cpuDiscovered CPUInfo

	cpuDiscovered.Discover()

	k := 0

	// Select cores that share LLC but do not share L1, L2 cache.
	// Return error if we can not provide cores meeting criterion.
	if filters == ShareLLCButNotL1L2 {

		// If count is more than cores per socket return error.
		if count == 0 || count > cpuDiscovered.PhysicalCores {
			return errors.New("Error - Insufficient cores")
		}

		for i := 0; i < cpuDiscovered.Sockets; i++ {
			k = 0

			for j := 0; j < cpuDiscovered.PhysicalCores; j++ {
				threadset.requestedThreads[k] = i*cpuDiscovered.PhysicalCores + j
				k++

				if k == count {
					break
				}
			}
			if k == count {
				break
			}
		}
	}
	if k < count {
		return errors.New("Error - Insufficient cores")
	}
	return nil
}
