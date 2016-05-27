package isolation

import (
	"errors"
)

// ShareLLCButNotL1L2 filter for selecting cpu.
const ShareLLCButNotL1L2 = 1 << iota

// CPUSelect with the desired criterion.
// Select cores that share LLC but do not share L1, L2 cache.
// Return error if we can not provide cores meeting criterion.
// Scan all the sockets for set of core ids meeting filter
func CPUSelect(countRequested int, filters uint) (IntSet, error) {

	threadSet := NewIntSet()
	// If countRequested is zero then return error.
	if countRequested == 0 {
		return nil, errors.New("Error - CPUSelect- number of core requested is zero")
	}

	// cpuDiscovered collect CPU topology.
	var cpuDiscovered CPUInfo

	err := cpuDiscovered.Discover()
	if err != nil {
		return nil, err
	}

	corecount := 0

	// Scan all the sockets for set of core ids meeting filter
	if filters == ShareLLCButNotL1L2 {

		// Return error if we can not provide cores meeting criterion.
		// If countRequested is more than cores per socket return error.
		if countRequested > cpuDiscovered.PhysicalCores {
			return nil, errors.New("Error - CPUSelect - insufficient cores")
		}
		// Loop through all the sockets first regular HW threads (lower ids) then Hyperthreads(upper ids)
		for socketid := 0; socketid < cpuDiscovered.Sockets*2; socketid++ {
			corecount = 0
			// Loop through all the cores to find available core ids meeting filter
			for cores := 0; cores < cpuDiscovered.PhysicalCores; cores++ {
				threadSet.Add(socketid*cpuDiscovered.PhysicalCores + cores)
				corecount++

				if corecount == countRequested {
					break
				}
			}
			if corecount == countRequested {
				break
			}
		}
	}
	if corecount < countRequested {
		return nil, errors.New("Error - CPUSelect - Insufficient cores to meet selection criteria")
	}
	return threadSet, nil
}
