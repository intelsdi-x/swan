package isolation

import (
	. "github.com/smartystreets/goconvey/convey"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

func TestDiscover(t *testing.T) {
	//	cpuTopo := CPUInfo{Sockets: 2, PhysicalCores: 4, ThreadsPerCore: 2, CacheL1i: 8192, CacheL1d: 8192, CacheL2: 262144, CacheL3: 8388608}
	cpuTopo := CPUInfo{Sockets: 1, PhysicalCores: 1, ThreadsPerCore: 1, CacheL1i: 1, CacheL1d: 1, CacheL2: 1, CacheL3: 1}

	Convey("Should provide Discover() to return and correct cpu topology", t, func() {
		So(cpuTopo.Discover(), ShouldBeNil)

		Convey("Should provide lscpu output without error", func() {
			out, err := exec.Command("lscpu").Output()
			So(err, ShouldBeNil)
			So(out, ShouldNotBeNil)

			outstring := strings.TrimSpace(string(out))
			lines := strings.Split(outstring, "\n")

			numaNodes := 0
			numCPUs := 0
			for _, line := range lines {
				fields := strings.Split(line, ":")
				if len(fields) < 2 {
					continue
				}
				key := strings.TrimSpace(fields[0])
				value := strings.TrimSpace(fields[1])

				switch key {
				case "NUMA node(s)":
					t, _ := strconv.Atoi(value)
					numaNodes = int(t)

				case "CPU(s)":
					t, _ := strconv.Atoi(value)
					numCPUs = int(t)
				}

			}
			So(err, ShouldBeNil)
			So(cpuTopo.Sockets, ShouldEqual, numaNodes)
			So(cpuTopo.PhysicalCores*cpuTopo.Sockets*cpuTopo.ThreadsPerCore, ShouldEqual, numCPUs)
		})
	})

}
