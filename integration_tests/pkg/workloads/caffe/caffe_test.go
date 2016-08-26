package integration

import (
	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/executor"
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/intelsdi-x/swan/pkg/workloads/caffe"
	"time"
)

func TestCaffeWithMockedExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("When I create Caffe with local executor and default configuration", t, func() {
		localExecutor := executor.NewLocal()
		c := caffe.New(localExecutor, caffe.DefaultConfig())

		Convey("When I launch the workload", func() {
			handle, err := c.Launch()
			defer handle.Stop()
			defer handle.Clean()
			defer handle.EraseOutput()

			Convey("Error is nil", func() {
				So(err, ShouldBeNil)

				Convey("Proper handle is returned", func() {
					So(handle, ShouldNotBeNil)

					Convey("Should work for at least one sec", func() {
						isTerminated := handle.Wait(1 * time.Second)
						So(isTerminated, ShouldBeFalse)

						Convey("Should be able to stop with no problem and be terminated", func() {
							err = handle.Stop()
							So(err, ShouldBeNil)

							state := handle.Status()
							So(state, ShouldEqual, executor.TERMINATED)
						})
					})
				})
			})
		})
	})
}
