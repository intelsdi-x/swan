package main

import (
	"testing"

	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/isolation/topo"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetSiblingThread(t *testing.T) {
	allThreads, err := topo.Discover()
	check(err)

	socket, err := allThreads.Sockets(1)
	check(err)

	if len(socket.AvailableThreads()) == len(socket.AvailableCores()) {
		t.Skipf("Cores do not seem to have hyper threading enabled. Skipping sibling test.")
	}

	if len(socket.AvailableCores()) < 2 {
		t.Skipf("Only one core available. skipping sibling test.")
	}

	coreThreads, err := socket.Cores(2)
	check(err)

	hpThreads, err := coreThreads.Threads(2)
	check(err)

	Convey("When obtaining siblings of hyperthread", t, func() {
		siblings := getSiblingThreadsOfThreadSet(hpThreads)

		Convey("Result siblings should be same size as entry threads and both should be nonempty", func() {
			So(hpThreads, ShouldNotBeEmpty)
			So(siblings, ShouldNotBeEmpty)
			So(len(hpThreads), ShouldEqual, len(siblings))
		})

		Convey("Result siblings and entry threads should each have at most one thread per core", func() {
			So(len(hpThreads.AvailableThreads()), ShouldEqual, len(hpThreads.AvailableCores()))
			So(len(siblings.AvailableThreads()), ShouldEqual, len(siblings.AvailableCores()))
		})

		Convey("Result siblings should be disjoint from entry threads", func() {
			So(hpThreads.AvailableThreads().Intersection(siblings.AvailableThreads()), ShouldResemble, isolation.NewIntSet())
		})

		Convey("Result siblings should reside on the same cores as entry threads", func() {
			So(hpThreads.AvailableCores().Equals(siblings.AvailableCores()), ShouldBeTrue)
		})
	})
}
