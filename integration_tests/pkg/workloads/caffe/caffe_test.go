package integration

import (
	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"fmt"
	"github.com/intelsdi-x/swan/pkg/workloads/caffe"
	"io/ioutil"
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
			Convey("Should work for at least one sec", func() {
				isTerminated := handle.Wait(1 * time.Second)
				So(isTerminated, ShouldBeFalse)
			})

			//Convey("Proper handle is returned", func() {
			//	So(handle, ShouldNotBeNil)
			//})
			//Convey("Error is nil", func() {
			//	So(err, ShouldBeNil)
			//})

			_ = err
			_ = handle

			file, err := handle.StderrFile()
			content, err := ioutil.ReadFile(file.Name())
			log.Error(fmt.Sprintf("%s", content))

		})
	})
}
