package topo

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/isolation/topo"
)

const lscpu1 = `# The following is the parsable format, which can be fed to other
# programs. Each different item in every column has an unique ID
# starting from zero.
# CPU,Core,Socket,Node,,L1d,L1i,L2,L3
0,0,0,0,,0,0,0,0
1,1,0,0,,1,1,1,1
`

const lscpu2 = `# The following is the parsable format, which can be fed to other
# programs. Each different item in every column has an unique ID
# starting from zero.
# CPU,Core,Socket,Node,,L1d,L1i,L2,L3
0,0,0,0,,0,0,0,0
1,1,0,0,,1,1,1,0
2,2,0,0,,2,2,2,0
3,3,0,0,,3,3,3,0
4,0,0,0,,0,0,0,0
5,1,0,0,,1,1,1,0
6,2,0,0,,2,2,2,0
7,3,0,0,,3,3,3,0
`

func TestDiscover(t *testing.T) {
	Convey("When discovering the CPU topology", t, func() {
		threadSet, err := topo.Discover()
		So(err, ShouldBeNil)

		Convey("It should discover a nonzero number of threads", func() {
			So(len(threadSet), ShouldBeGreaterThan, 0)
		})
	})
}

func TestReadTopology(t *testing.T) {
	Convey("When reading a canned topology with 2 CPUs", t, func() {
		threadSet, err := topo.ReadTopology([]byte(lscpu1))
		So(err, ShouldBeNil)

		Convey("It should return a thread set with 2 threads from 2 cores on the same socket", func() {
			So(len(threadSet), ShouldEqual, 2)
			So(threadSet.AvailableThreads().Equals(isolation.NewIntSet(0, 1)), ShouldBeTrue)
			So(threadSet.AvailableCores().Equals(isolation.NewIntSet(0, 1)), ShouldBeTrue)
			So(threadSet.AvailableSockets().Equals(isolation.NewIntSet(0)), ShouldBeTrue)
		})
	})

	Convey("When reading a canned topology with 8 CPUs", t, func() {
		threadSet, err := topo.ReadTopology([]byte(lscpu2))
		So(err, ShouldBeNil)

		Convey("It should return a thread set with 8 threads from 4 cores on 1 socket", func() {
			So(len(threadSet), ShouldEqual, 8)
			So(threadSet.AvailableThreads().Equals(isolation.NewIntSet(0, 1, 2, 3, 4, 5, 6, 7)), ShouldBeTrue)
			So(threadSet.AvailableCores().Equals(isolation.NewIntSet(0, 1, 2, 3)), ShouldBeTrue)
			So(threadSet.AvailableSockets().Equals(isolation.NewIntSet(0)), ShouldBeTrue)
		})
	})
}
