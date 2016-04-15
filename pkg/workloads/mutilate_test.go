package workloads

import (
	log "github.com/Sirupsen/logrus"
	//"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/executor"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestTuneMutilate(t *testing.T) {
	Convey("When Tuning Mutilate to Memcached.", t, func() {
		localExecutor := executor.NewLocal()

		memcached_uri := "localhost"
		mutilate_path := "/home/skonefal/dev/src/mutilate/mutilate"

		mutilate := NewMutilate(localExecutor, memcached_uri, mutilate_path)

		slo := 1000
		percentile := 50

		targetQps, err := mutilate.Tune(slo, percentile)

		//targetQps := 0
		//var err error

		_ = mutilate
		_ = slo
		_ = percentile

		Convey("Error should be nil.", func() {
			So(err, ShouldBeNil)
		})
		Convey("TargetQPS should be more than 0.", func() {
			So(targetQps, ShouldNotEqual, 0)
		})
	})
}

func TestGetTuningOutput(t *testing.T) {
	log.SetLevel(log.ErrorLevel)
	const correctMutilateOutput = `#type       avg     std     min     5th    10th    90th    95th    99th
read       20.9    11.9    11.9    12.5    13.1    32.4    39.0    56.8
update      0.0     0.0     0.0     0.0     0.0     0.0     0.0     0.0
op_q        1.0     0.0     1.0     1.0     1.0     1.1     1.1     1.1

Total QPS = 4450.3 (89007 / 20.0s)
Peak QPS  = 71164.8

Misses = 0 (0.0%)
Skipped TXs = 0 (0.0%)

RX   22044729 bytes :    1.1 MB/s
TX    3204252 bytes :    0.2 MB/s
`
	const correctTargetQps = 4450

	Convey("When given proper Mutilate tuning output: ", t, func() {
		targetQps, err := getTargetQps(correctMutilateOutput)

		Convey("we receive proper targetQPS.", func() {
			So(err, ShouldBeNil)
		})

		Convey("we receive nil error.", func() {
			So(targetQps, ShouldEqual, correctTargetQps)
		})
	})

	Convey("When given inproper Mutilate tuning output we receive error.", t, func() {
		_, err := getTargetQps("Inproper Output")
		So(err, ShouldNotBeNil)
	})
}
