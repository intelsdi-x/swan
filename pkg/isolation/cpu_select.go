package isolation

import (
	"errors"
)

// ShareLLCButNotL1L2 filter for selecting cpu.
const ShareLLCButNotL1L2 uint = 1

// CPUSelect with the desired criterion.
// Select cores that share LLC but do not share L1, L2 cache.
// Return error if we can not provide cores meeting criterion.
// Scan all the sockets for set of core ids meeting filter
func (threadset *IntSet) CPUSelect(count int, filters uint) error {

	// cpuDiscovered collect CPU topology.
	var cpuDiscovered CPUInfo

	err := cpuDiscovered.Discover()
	if err != nil {
		return err
	}

	corecount := 0

	// Scan all the sockets for set of core ids meeting filter
	if filters == ShareLLCButNotL1L2 {

		// Return error if we can not provide cores meeting criterion.
		// If count is more than cores per socket return error.
		if count > cpuDiscovered.PhysicalCores {
			return errors.New("Error - CPUSelect - insufficient cores")
		}
		// If count is more than cores per socket return error.
		if count == 0 {
			return errors.New("Error - CPUSelect- number of core requested is zero")
		}
		// Loop through all the sockets
		for socketid := 0; socketid < cpuDiscovered.Sockets*2; socketid++ {
			corecount = 0
			// Loop through all the cores to find available core ids meeting filter
			for cores := 0; cores < cpuDiscovered.PhysicalCores; cores++ {
				threadset.Add(socketid*cpuDiscovered.PhysicalCores + cores)
				corecount++

				if corecount == count {
					break
				}
			}
			if corecount == count {
				break
			}
		}
	}
	if corecount < count {
		return errors.New("Error - CPUSelect - Insufficient cores to meet selection criteria")
	}
	return nil
}
