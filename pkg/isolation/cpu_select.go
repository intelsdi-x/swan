package isolation

import (
	"errors"
)

// ShareLLCButNotL1L2 filter for selecting cpu.
const ShareLLCButNotL1L2 uint = 1

// CPUSelect with the desired criterion.
func (threadset *IntSet) CPUSelect(count int, filters uint) error {

	// cpuDiscovered collect CPU topology.
	var cpuDiscovered CPUInfo

	err := cpuDiscovered.Discover()
	if err != nil {
		return err
	}

	k := 0
	// Select cores that share LLC but do not share L1, L2 cache.
	// Return error if we can not provide cores meeting criterion.
	if filters == ShareLLCButNotL1L2 {

		// If count is more than cores per socket return error.
		if count > cpuDiscovered.PhysicalCores {
			return errors.New("Error - Insufficient cores")
		}
		// If count is more than cores per socket return error.
		if count == 0 {
			return errors.New("Error - Number of core requested is zero")
		}

		for i := 0; i < cpuDiscovered.Sockets*2; i++ {
			k = 0

			for j := 0; j < cpuDiscovered.PhysicalCores; j++ {
				threadset.Add(i*cpuDiscovered.PhysicalCores + j)
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
		return errors.New("Error - Insufficient cores to meet selection criteria")
	}
	return nil
}
