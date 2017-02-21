package integration

import (
	"fmt"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/intelsdi-x/swan/pkg/workloads/caffe"
)

func TestCaffeWithMockedExecutor(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	Convey("When I create Caffe with local executor and default configuration", t, func() {
		localExecutor := executor.NewLocal()
		caffeConfig := caffe.DefaultConfig()
		fmt.Printf("caffeConfig = %+v\n", caffeConfig)
		c := caffe.New(localExecutor, caffeConfig)

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
