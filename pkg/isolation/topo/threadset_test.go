package topo

import (
	"testing"

	"github.com/intelsdi-x/athena/pkg/isolation"
	. "github.com/smartystreets/goconvey/convey"
)

func syntheticThreadSet() ThreadSet {
	return ThreadSet{
		// id, core, socket
		NewThread(1, 1, 1),
		NewThread(2, 1, 1),
		NewThread(3, 2, 1),
		NewThread(4, 2, 1),
		NewThread(5, 3, 2),
		NewThread(6, 3, 2),
		NewThread(7, 4, 2),
		NewThread(8, 4, 2),
	}
}

func TestNewThreadSet(t *testing.T) {
	Convey("When creating a new thread set", t, func() {
		ts := NewThreadSet()
		Convey("It should be empty", func() {
			So(len(ts), ShouldEqual, 0)
		})
	})
}

func TestThreadSetAvailableThreads(t *testing.T) {
	Convey("Given a synthetic thread set", t, func() {
		ts := syntheticThreadSet()
		Convey("Listing the available threads should yield a set of size 8", func() {
			So(ts.AvailableThreads().Equals(isolation.NewIntSet(1, 2, 3, 4, 5, 6, 7, 8)), ShouldBeTrue)
		})
	})
}

func TestThreadSetAvailableCores(t *testing.T) {
	Convey("Given a synthetic thread set", t, func() {
		ts := syntheticThreadSet()
		Convey("Listing the available cores should yield a set of size 4", func() {
			So(ts.AvailableCores().Equals(isolation.NewIntSet(1, 2, 3, 4)), ShouldBeTrue)
		})
	})
}

func TestThreadSetAvailableSockets(t *testing.T) {
	Convey("Given a synthetic thread set", t, func() {
		ts := syntheticThreadSet()
		Convey("Listing the available sockets should yield a set of size 2", func() {
			So(ts.AvailableSockets().Equals(isolation.NewIntSet(1, 2)), ShouldBeTrue)
		})
	})
}

func TestThreadSetPartition(t *testing.T) {
	Convey("Given a synthetic thread set", t, func() {
		ts := syntheticThreadSet()
		Convey("Partitioning thread ids into evens and odds should equal thread sets of equal length", func() {
			evens, odds := ts.Partition(func(t Thread) bool { return t.ID()%2 == 0 })
			So(len(evens), ShouldEqual, len(odds))
			So(evens.AvailableThreads().Equals(isolation.NewIntSet(2, 4, 6, 8)), ShouldBeTrue)
			So(odds.AvailableThreads().Equals(isolation.NewIntSet(1, 3, 5, 7)), ShouldBeTrue)
		})
	})
}

func TestThreadSetFilter(t *testing.T) {
	Convey("Given a synthetic thread set", t, func() {
		ts := syntheticThreadSet()
		Convey("Filtering out the odd thread ids should yield only the evens", func() {
			evens := ts.Filter(func(t Thread) bool { return t.ID()%2 == 0 })
			So(len(evens), ShouldEqual, 4)
			So(evens.AvailableThreads().Equals(isolation.NewIntSet(2, 4, 6, 8)), ShouldBeTrue)
		})
	})
}

func TestThreadSetThreads(t *testing.T) {
	Convey("Given a synthetic thread set", t, func() {
		ts := syntheticThreadSet()
		Convey("Requesting 3 threads should yield 3 disjoint threads", func() {
			threads, err := ts.Threads(3)
			So(err, ShouldBeNil)
			So(len(threads), ShouldEqual, 3)
			So(len(threads.AvailableThreads()), ShouldEqual, 3)
		})

		Convey("Requesting 10 threads should yield an error", func() {
			threads, err := ts.Threads(10)
			So(err, ShouldNotBeNil)
			So(threads, ShouldBeNil)
		})
	})
}

func TestThreadSetCores(t *testing.T) {
	Convey("Given a synthetic thread set", t, func() {
		ts := syntheticThreadSet()
		Convey("Requesting 2 cores should yield 4 threads from 2 cores", func() {
			threads, err := ts.Cores(2)
			So(err, ShouldBeNil)
			So(len(threads), ShouldEqual, 4)
			So(len(threads.AvailableThreads()), ShouldEqual, 4)
			So(len(threads.AvailableCores()), ShouldEqual, 2)
		})

		Convey("Requesting 4 cores should yield 8 threads from 4 cores", func() {
			threads, err := ts.Cores(4)
			So(err, ShouldBeNil)
			So(len(threads), ShouldEqual, 8)
			So(len(threads.AvailableThreads()), ShouldEqual, 8)
			So(len(threads.AvailableCores()), ShouldEqual, 4)
		})

		Convey("Requesting 5 cores should yield an error", func() {
			threads, err := ts.Cores(5)
			So(err, ShouldNotBeNil)
			So(threads, ShouldBeNil)
		})
	})
}

func TestThreadSetSockets(t *testing.T) {
	Convey("Given a synthetic thread set", t, func() {
		ts := syntheticThreadSet()
		Convey("Requesting 1 socket should yield 4 threads from 2 cores and 1 socket", func() {
			threads, err := ts.Sockets(1)
			So(err, ShouldBeNil)
			So(len(threads), ShouldEqual, 4)
			So(len(threads.AvailableThreads()), ShouldEqual, 4)
			So(len(threads.AvailableCores()), ShouldEqual, 2)
			So(len(threads.AvailableSockets()), ShouldEqual, 1)
		})

		Convey("Requesting 2 sockets should yield 8 threads from 4 cores and 2 sockets", func() {
			threads, err := ts.Sockets(2)
			So(err, ShouldBeNil)
			So(len(threads), ShouldEqual, 8)
			So(len(threads.AvailableThreads()), ShouldEqual, 8)
			So(len(threads.AvailableCores()), ShouldEqual, 4)
			So(len(threads.AvailableSockets()), ShouldEqual, 2)
		})

		Convey("Requesting 3 sockets should yield an error", func() {
			threads, err := ts.Sockets(3)
			So(err, ShouldNotBeNil)
			So(threads, ShouldBeNil)
		})
	})
}

func TestThreadSetFromThreads(t *testing.T) {
	Convey("Given a synthetic thread set", t, func() {
		ts := syntheticThreadSet()
		Convey("Requesting threads with IDs 1, 2, 5 and 6 should yield 4 threads from 2 cores and 2 sockets", func() {
			threads, err := ts.FromThreads(1, 2, 5, 6)
			So(err, ShouldBeNil)
			So(len(threads), ShouldEqual, 4)
			So(threads.AvailableThreads().Equals(isolation.NewIntSet(1, 2, 5, 6)), ShouldBeTrue)
			So(threads.AvailableCores().Equals(isolation.NewIntSet(1, 3)), ShouldBeTrue)
			So(threads.AvailableSockets().Equals(isolation.NewIntSet(1, 2)), ShouldBeTrue)
		})

		Convey("Requesting threads with IDs 1, 3, and 9 should yield an error", func() {
			threads, err := ts.FromThreads(1, 3, 9)
			So(err, ShouldNotBeNil)
			So(threads, ShouldBeNil)
		})
	})
}

func TestThreadSetFromCores(t *testing.T) {
	Convey("Given a synthetic thread set", t, func() {
		ts := syntheticThreadSet()
		Convey("Requesting threads from cores 1 and 3 should yield 4 threads from 2 cores and 2 sockets", func() {
			threads, err := ts.FromCores(1, 3)
			So(err, ShouldBeNil)
			So(len(threads), ShouldEqual, 4)
			So(threads.AvailableThreads().Equals(isolation.NewIntSet(1, 2, 5, 6)), ShouldBeTrue)
			So(threads.AvailableCores().Equals(isolation.NewIntSet(1, 3)), ShouldBeTrue)
			So(threads.AvailableSockets().Equals(isolation.NewIntSet(1, 2)), ShouldBeTrue)
		})

		Convey("Requesting threads from cores 1, 3, and 5 should yield an error", func() {
			threads, err := ts.FromCores(1, 3, 5)
			So(err, ShouldNotBeNil)
			So(threads, ShouldBeNil)
		})
	})
}

func TestThreadSetFromSockets(t *testing.T) {
	Convey("Given a synthetic thread set", t, func() {
		ts := syntheticThreadSet()
		Convey("Requesting threads from socket 2 should yield 4 threads from 2 cores and 1 sockets", func() {
			threads, err := ts.FromSockets(2)
			So(err, ShouldBeNil)
			So(len(threads), ShouldEqual, 4)
			So(threads.AvailableThreads().Equals(isolation.NewIntSet(5, 6, 7, 8)), ShouldBeTrue)
			So(threads.AvailableCores().Equals(isolation.NewIntSet(3, 4)), ShouldBeTrue)
			So(threads.AvailableSockets().Equals(isolation.NewIntSet(2)), ShouldBeTrue)
		})

		Convey("Requesting threads from both sockets should yield 8 threads from 4 cores and 2 sockets", func() {
			threads, err := ts.FromSockets(2, 1)
			So(err, ShouldBeNil)
			So(len(threads), ShouldEqual, 8)
			So(threads.AvailableThreads().Equals(isolation.NewIntSet(1, 2, 3, 4, 5, 6, 7, 8)), ShouldBeTrue)
			So(threads.AvailableCores().Equals(isolation.NewIntSet(1, 2, 3, 4)), ShouldBeTrue)
			So(threads.AvailableSockets().Equals(isolation.NewIntSet(1, 2)), ShouldBeTrue)
		})

		Convey("Requesting threads from sockets 1 and 3 should yield an error", func() {
			threads, err := ts.FromSockets(1, 3)
			So(err, ShouldNotBeNil)
			So(threads, ShouldBeNil)
		})
	})
}

func TestThreadSetContains(t *testing.T) {
	Convey("Given a synthetic thread set", t, func() {
		ts := syntheticThreadSet()
		Convey("It should contain all of its threads", func() {
			for _, th := range ts {
				So(ts.Contains(th), ShouldBeTrue)
			}
		})
		Convey("And it should not contain other threads", func() {
			So(ts.Contains(NewThread(0, 0, 0)), ShouldBeFalse)
		})
	})
}

func TestThreadSetDifference(t *testing.T) {
	Convey("Given a synthetic thread set", t, func() {
		ts := syntheticThreadSet()
		Convey("Subtracting all threads from socket 1 should yield 4 threads from 2 cores and 1 sockets", func() {
			s1Threads, err := ts.FromSockets(1)
			So(err, ShouldBeNil)

			s2Threads := ts.Difference(s1Threads)

			s2ThreadsExpected, err := ts.FromSockets(2)
			So(err, ShouldBeNil)

			So(len(s2Threads), ShouldEqual, 4)
			So(s2Threads, ShouldResemble, s2ThreadsExpected)
			So(s2Threads.AvailableThreads().Equals(isolation.NewIntSet(5, 6, 7, 8)), ShouldBeTrue)
			So(s2Threads.AvailableCores().Equals(isolation.NewIntSet(3, 4)), ShouldBeTrue)
			So(s2Threads.AvailableSockets().Equals(isolation.NewIntSet(2)), ShouldBeTrue)
		})
	})
}

func TestThreadSetRemoveThread(t *testing.T) {
	Convey("Given a synthetic thread set", t, func() {
		ts := syntheticThreadSet()
		Convey("Removing all threads from socket 1 should yield 4 threads from 2 cores and 1 sockets", func() {
			s1Threads, err := ts.FromSockets(1)
			So(err, ShouldBeNil)

			s2Threads := ts
			for _, th := range s1Threads {
				s2Threads = s2Threads.Remove(th)
			}

			s2ThreadsExpected, err := ts.FromSockets(2)
			So(err, ShouldBeNil)

			So(len(s2Threads), ShouldEqual, 4)
			So(s2Threads, ShouldResemble, s2ThreadsExpected)
			So(s2Threads.AvailableThreads().Equals(isolation.NewIntSet(5, 6, 7, 8)), ShouldBeTrue)
			So(s2Threads.AvailableCores().Equals(isolation.NewIntSet(3, 4)), ShouldBeTrue)
			So(s2Threads.AvailableSockets().Equals(isolation.NewIntSet(2)), ShouldBeTrue)
		})
	})
}

func TestThreadSetRemoveThreadSet(t *testing.T) {
	Convey("Given a synthetic thread set", t, func() {
		ts := syntheticThreadSet()
		Convey("Removing all threads from socket 1 should yield 4 threads from 2 cores and 1 sockets", func() {
			s1Threads, err := ts.FromSockets(1)
			So(err, ShouldBeNil)

			s2Threads := ts.RemoveThreadSet(s1Threads)

			s2ThreadsExpected, err := ts.FromSockets(2)
			So(err, ShouldBeNil)

			So(len(s2Threads), ShouldEqual, 4)
			So(s2Threads, ShouldResemble, s2ThreadsExpected)
			So(s2Threads.AvailableThreads().Equals(isolation.NewIntSet(5, 6, 7, 8)), ShouldBeTrue)
			So(s2Threads.AvailableCores().Equals(isolation.NewIntSet(3, 4)), ShouldBeTrue)
			So(s2Threads.AvailableSockets().Equals(isolation.NewIntSet(2)), ShouldBeTrue)
		})
	})
}

func TestThreadSetToCpuSet(t *testing.T) {
	Convey("Given a synthetic thread set", t, func() {
		ts := syntheticThreadSet()
		Convey("Written in CpuSet notation it should be `1,2,3,4,5,6,7,8`", func() {
			cpuset := ts.ToCPUSetNotation()
			So(cpuset, ShouldEqual, "1,2,3,4,5,6,7,8")
		})
	})
}
