package isolation

import "os/exec"
import "strconv"
import "strings"
import "errors"
import "github.com/pivotal-golang/bytefmt"

// CPUInfo defines data needed for CPU topology
type CPUInfo struct {
	sockets        int
	physicalCores  int
	threadsPerCore int
	cacheL1i       int
	cacheL1d       int
	cacheL2        int
	cacheL3        int
}

// NewCPUInfo instance creation.
func NewCPUInfo(cores int, threads int, l1i int, l1d int, l2 int, l3 int) *CPUInfo {
	return &CPUInfo{physicalCores: cores, threadsPerCore: threads, cacheL1i: l1i, cacheL1d: l1d, cacheL2: l2, cacheL3: l3}
}

// Discover CPU topology and caches sizes. We use lscpu for the time being until we can make this code more portable by reading directly from HW
func (cputopo *CPUInfo) Discover() error {

	out, err := exec.Command("lscpu").Output()
	if err != nil {
		return err
	}

	outstring := strings.TrimSpace(string(out))
	lines := strings.Split(outstring, "\n")

	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])

		switch key {
		case "Socket(s)":
			t, err := strconv.Atoi(value)
			cputopo.sockets = int(t)
			if err != nil {
				return errors.New("Unexpected input error")
			}
		case "Core(s) per socket":
			t, err := strconv.Atoi(value)
			cputopo.physicalCores = int(t)
			if err != nil {
				return errors.New("Unexpected input error")
			}
		case "Thread(s) per core":
			t, err := strconv.Atoi(value)
			cputopo.threadsPerCore = int(t)
			if err != nil {
				return errors.New("Unexpected input error")
			}
		case "L1d cache":
			mydata, err := bytefmt.ToBytes(value)
			cputopo.cacheL1d = int(mydata)
			if err != nil {
				return errors.New("Unexpected input error")
			}
		case "L1i cache":
			mydata, err := bytefmt.ToBytes(value)
			cputopo.cacheL1i = int(mydata)
			if err != nil {
				return errors.New("Unexpected input error")
			}
		case "L2 cache":
			mydata, err := bytefmt.ToBytes(value)
			cputopo.cacheL1i = int(mydata)
			if err != nil {
				return errors.New("Unexpected input error")
			}
		case "L3 cache":
			mydata, err := bytefmt.ToBytes(value)
			cputopo.cacheL3 = int(mydata)
			if err != nil {
				return errors.New("Unexpected input error")
			}

		}

	}

	return nil
}
