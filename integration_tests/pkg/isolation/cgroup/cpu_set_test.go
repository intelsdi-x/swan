// +build parallel

package integration

import (
	"github.com/intelsdi-x/athena/pkg/isolation"
	"os/exec"
	pth "path"
	"testing"

	"github.com/intelsdi-x/athena/pkg/isolation/cgroup"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCPUSet(t *testing.T) {

	Convey("When constructing a new CPUSet", t, func() {
		uuid1 := uuidgen(t)
		uuid2 := uuidgen(t)
		uuid3 := uuidgen(t)

		path := pth.Join("/", uuid1, uuid2, uuid3)
		cpus := isolation.NewIntSet(1)
		mems := isolation.NewIntSet(0)

		// Setting these to true assumes too much about the environment...
		// For example the docker cpuset cgroup assigns all cpus by default,
		// meaning any other cpuset overlaps, precluding exclusivity.
		cpuExclusive := false
		memExclusive := false

		Convey("Creating the CPUSet should create and set the correct values for the underlying Cgroup", func() {
			cpuSet, err := cgroup.NewCPUSet(path, cpus, mems, cpuExclusive, memExclusive)
			So(err, ShouldBeNil)

			So(cpuSet.Cgroup().Path(), ShouldEqual, path)
			So(cpuSet.Cpus(), ShouldEqual, cpus)
			So(cpuSet.Mems(), ShouldEqual, mems)
			So(cpuSet.CPUExclusive(), ShouldEqual, cpuExclusive)
			So(cpuSet.MemExclusive(), ShouldEqual, memExclusive)

			So(cpuSet.Create(), ShouldBeNil)
			defer cpuSet.Cgroup().Ancestors()[1].Destroy(true)

			exists, err := cpuSet.Cgroup().Exists()
			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)

			actual, err := cpuSet.Cgroup().Get(cgroup.CPUSetCpus)
			So(err, ShouldBeNil)
			set, err := isolation.NewIntSetFromRange(actual)
			So(err, ShouldBeNil)
			So(set.Equals(cpus), ShouldBeTrue)

			actual, err = cpuSet.Cgroup().Get(cgroup.CPUSetMems)
			So(err, ShouldBeNil)
			set, err = isolation.NewIntSetFromRange(actual)
			So(err, ShouldBeNil)
			So(set.Equals(mems), ShouldBeTrue)

			actual, err = cpuSet.Cgroup().Get(cgroup.CPUSetCPUExclusive)
			So(err, ShouldBeNil)

			err = cpuSet.Cgroup().SetAndCheck(cgroup.CPUSetCPUExclusive, "1")
			So(actual, ShouldEqual, "0")

			actual, err = cpuSet.Cgroup().Get(cgroup.CPUSetMemExclusive)
			So(err, ShouldBeNil)
			So(actual, ShouldEqual, "0")
		})

		Convey("After starting an unisolated process", func() {
			cpuSet, err := cgroup.NewCPUSet(path, cpus, mems, cpuExclusive, memExclusive)
			So(err, ShouldBeNil)
			So(cpuSet.Create(), ShouldBeNil)
			defer cpuSet.Cgroup().Ancestors()[1].Destroy(true)

			cmd := exec.Command("sleep", "300") // 5 minutes
			err = cmd.Start()
			defer cmd.Process.Kill()
			So(err, ShouldBeNil)

			Convey("It should isolate the process", func() {
				So(cpuSet.Isolate(cmd.Process.Pid), ShouldBeNil)
				tasks, err := cpuSet.Cgroup().Tasks(cgroup.CPUSetController)
				So(err, ShouldBeNil)
				So(tasks.Contains(cmd.Process.Pid), ShouldBeTrue)
			})
		})

		Convey("It should clean up after itself", func() {
			cpuSet, err := cgroup.NewCPUSet(path, cpus, mems, cpuExclusive, memExclusive)
			So(err, ShouldBeNil)
			So(cpuSet.Create(), ShouldBeNil)
			defer cpuSet.Cgroup().Ancestors()[1].Destroy(true)

			So(cpuSet.Clean(), ShouldBeNil)
			exists, err := cpuSet.Cgroup().Exists()
			So(err, ShouldBeNil)
			So(exists, ShouldBeFalse)
		})
	})
}
