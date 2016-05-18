package isolation

import "os/exec"
import "strconv"
import "strings"
import "errors"

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

// const for human readable byte sizes
const (
	BYTE     = 1.0
	KILOBYTE = 1024 * BYTE
	MEGABYTE = 1024 * KILOBYTE
	GIGABYTE = 1024 * MEGABYTE
	TERABYTE = 1024 * GIGABYTE
)

// NewCPUInfo instance creation.
func NewCPUInfo(cores int, threads int, l1i int, l1d int, l2 int, l3 int) *CPUInfo {
	return &CPUInfo{physicalCores: cores, threadsPerCore: threads, cacheL1i: l1i, cacheL1d: l1d, cacheL2: l2, cacheL3: l3}
}

// UnitsOfBytes returns KILOBYTES, MEGABYTES, etc as detected in the input string
func UnitsOfBytes(s string) (int, error) {
	if strings.Contains(s, "K") {
		return KILOBYTE, nil
	} else if strings.Contains(s, "M") {
		return MEGABYTE, nil
	} else if strings.Contains(s, "G") {
		return GIGABYTE, nil
	} else if strings.Contains(s, "T") {
		return TERABYTE, nil
	} else {
		err := errors.New("Unexpected input error")
		return BYTE, err
	}
}

// Discover CPU topology and caches sizes
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
			t, _ := strconv.Atoi(value)
			cputopo.sockets = int(t)
		case "Core(s) per socket":
			t, _ := strconv.Atoi(value)
			cputopo.physicalCores = int(t)
		case "Thread(s) per core":
			t, _ := strconv.Atoi(value)
			cputopo.threadsPerCore = int(t)
		case "L1d cache":
			tFmt := value[:len(value)-1]
			t, _ := strconv.Atoi(tFmt)

			multiplier, err := UnitsOfBytes(value)
			if err == nil {
				cputopo.cacheL1d = int(t) * multiplier
			}
		case "L1i cache":
			tFmt := value[:len(value)-1]
			t, _ := strconv.Atoi(tFmt)

			multiplier, err := UnitsOfBytes(value)
			if err == nil {
				cputopo.cacheL1i = int(t) * multiplier
			}
		case "L2 cache":
			tFmt := value[:len(value)-1]
			t, _ := strconv.Atoi(tFmt)

			multiplier, err := UnitsOfBytes(value)
			if err == nil {
				cputopo.cacheL2 = int(t) * multiplier
			}
		case "L3 cache":
			tFmt := value[:len(value)-1]
			t, _ := strconv.Atoi(tFmt)

			multiplier, err := UnitsOfBytes(value)
			if err == nil {
				cputopo.cacheL3 = int(t) * multiplier
			}

		}

	}

	if err != nil {
		return err
	}

	return nil
}
