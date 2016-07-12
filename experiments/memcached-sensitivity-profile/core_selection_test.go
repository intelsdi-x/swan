package main

import (
	"testing"

	"github.com/intelsdi-x/swan/pkg/isolation/topo"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetSiblingThread(t *testing.T) {
	allThreads, err := topo.Discover()
	errutil.Check(err)

	socket, err := allThreads.Sockets(1)
	errutil.Check(err)

	if len(socket.AvailableThreads()) == len(socket.AvailableCores()) {
		t.Skipf("Cores does not seem to have hyper threading enabled. Skipping sibling test.")
	}

	hpThreads, err := socket.Threads(2)
	errutil.Check(err)

	Convey("When obtaining siblings of hyperthread", t, func() {
		siblings := getSiblingThreadsOfThreadSet(hpThreads)

		Convey("Result siblings should be same size as entry threads", func() {
			So(len(hpThreads), ShouldEqual, len(siblings))
		})

		Convey("Result siblings should be different than entry threads", func() {
			So(hpThreads, ShouldNotResemble, siblings)
		})
	})
}
