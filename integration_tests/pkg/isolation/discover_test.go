package isolation

import (
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/intelsdi-x/swan/pkg/isolation"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDiscover(t *testing.T) {
	cpuTopo := isolation.NewCPUInfo(1, 1, 1, 1, 1, 1, 1)

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
