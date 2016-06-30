package main

import (
	"testing"

	"github.com/intelsdi-x/swan/pkg/isolation/topo"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetSiblingThread(t *testing.T) {
	allThreads, err := topo.Discover()
	check(err)

	socket, err := allThreads.Sockets(1)
	check(err)

	hpThreads, err := socket.Threads(2)
	check(err)

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
