package isolation

import (
	. "github.com/smartystreets/goconvey/convey"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

func TestDiscover(t *testing.T) {
	cpuTopo := CPUInfo{physicalCores: 4, threadsPerCore: 2, cacheL1i: 8192, cacheL1d: 8192, cacheL2: 262144, cacheL3: 8388608}

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
			So(cpuTopo.sockets, ShouldEqual, numaNodes)
			So(cpuTopo.physicalCores*cpuTopo.sockets*cpuTopo.threadsPerCore, ShouldEqual, numCPUs)
		})
	})

}
