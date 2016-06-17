package isolation

import (
	"github.com/intelsdi-x/swan/pkg/isolation"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCPUSelect(t *testing.T) {

	// cpuDiscovered collect CPU topology.
	var cpus isolation.CPUInfo
	blackList := isolation.NewIntSet()

	cpus.Discover()

	Convey("Should build reasonable topology mappings", t, func() {
		So(len(cpus.SocketCores), ShouldEqual, cpus.Sockets)
		So(len(cpus.CoreCpus), ShouldEqual, cpus.PhysicalCores)

		numCpus := 0
		for _, cpus := range cpus.CoreCpus {
			numCpus += len(cpus)
		}
		So(numCpus, ShouldEqual, cpus.PhysicalCores*cpus.ThreadsPerCore)
	})

	Convey("Should provide CPUSelect() to return an error when requesting zero cpus", t, func() {
		threadset, err := isolation.CPUSelect(0, isolation.ShareLLCButNotL1L2, blackList)
		So(err, ShouldNotBeNil)

		Convey("Should have length zero", func() {
			So(threadset, ShouldHaveLength, 0)
		})
	})

	Convey("Should provide CPUSelect() to return nil and correct cpu ids", t, func() {
		threadset, err := isolation.CPUSelect(cpus.PhysicalCores, isolation.ShareLLCButNotL1L2, blackList)
		So(err, ShouldBeNil)

		Convey("Should have length", func() {
			So(threadset, ShouldHaveLength, cpus.PhysicalCores)
		})

		Convey("It should contain exactly one CPU from each physical core", func() {
			for core := range cpus.CoreCpus {
				So(len(threadset.Intersection(cpus.CoreCpus[core])), ShouldEqual, 1)
			}
		})
	})

	Convey("Should provide CPUSelect() to not return nil when requesting more cores than a socket has", t, func() {
		threadset, err := isolation.CPUSelect(cpus.PhysicalCores+1, isolation.ShareLLCButNotL1L2, blackList)
		So(err, ShouldNotBeNil)

		Convey("Should have length zero", func() {
			So(threadset, ShouldHaveLength, 0)
		})
	})

}
