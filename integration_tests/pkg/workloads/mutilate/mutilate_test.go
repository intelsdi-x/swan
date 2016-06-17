package mutilate

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/misc/snap-plugin-collector-mutilate/mutilate/parse"
	"github.com/intelsdi-x/swan/pkg/executor"
	fs "github.com/intelsdi-x/swan/pkg/utils/os"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	"github.com/shopspring/decimal"
	. "github.com/smartystreets/goconvey/convey"
)

// TestMutilateWithExecutor is an integration test with local executor.
// Flow:
// - start memcached and make sure it is a new clean instance
// - run populate and check new items were stored
// - run tune - search for capacity and ensure it is not zero
// - run load and check it run without error (ignore results)
// note: for Populate/Tune we don't check output files
func TestMutilateWithExecutor(t *testing.T) {
	// log.SetLevel(log.ErrorLevel)
	// log.SetOutput(os.Stderr)

	// memcached setup
	mcConfig := memcached.DefaultMemcachedConfig()
	mcConfig.User = fs.GetEnvOrDefault("USER", mcConfig.User)
	mcAddress := fmt.Sprintf("127.0.0.1:%d", mcConfig.Port)
	// closure function with memcached address for getting statistics from mecache

	// start memcached and make sure it is a new one!
	memcachedLauncher := memcached.New(executor.NewLocal(), mcConfig)
	mcHandle, err := memcachedLauncher.Launch()
	if err != nil {
		t.Fatal("cannot start memcached:" + err.Error())
	}

	// clean memacache
	defer func() {
		// and our memcached instance was closed properlly
		if err := mcHandle.Stop(); err != nil {
			t.Fatal(err)
		}
		mcHandle.Wait(0)
		if ec, err := mcHandle.ExitCode(); err != nil || ec != -1 {
			// expected -1 on SIGKILL (TODO: change to zero, after Stop "gracefull stop fix"
			t.Fatalf("memcached was stopped incorrectly err %s exit-code: %d", err, ec)
		}
		// make sure temp files removal was successful
		if err := mcHandle.EraseOutput(); err != nil {
			t.Fatal(err)
		}
	}()

	// Give memcached chance to start and possibly die
	if stopped := mcHandle.Wait(1 * time.Second); stopped {
		t.Fatal("memcached is not running after the second")
	}

	if mcHandle.Status() != executor.RUNNING {
		t.Fatal("memcached is not running!")
	}
	currItems, _ := getMemcachedStats(mcAddress, t)
	if currItems != 0 { // in case of not memached or someone at the same time is messing with it
		t.Fatal("expecting empty memcached but some items are already there")
	}

	Convey("With memcached started already", t, func() {

		mutilateConfig := mutilate.DefaultMutilateConfig()
		mutilateConfig.TuningTime = 1 * time.Second
		mutilateConfig.EraseSearchTuneOutput = true                       // make sure files are removed correctly
		mutilateConfig.LatencyPercentile, _ = decimal.NewFromString("99") // not sure if custom percentile is working correctly TODO: added a custom percentile integration test

		Convey("When run mutilate populate", func() {
			mutilateLauncher := mutilate.New(executor.NewLocal(), mutilateConfig)
			err := mutilateLauncher.Populate()
			So(err, ShouldBeNil)
			currItems, _ = getMemcachedStats(mcAddress, t)
			So(currItems, ShouldNotEqual, 0)
		})

		Convey("When run mutilate tune", func() {
			_, previousGetCnt := getMemcachedStats(mcAddress, t)
			mutilateLauncher := mutilate.New(executor.NewLocal(), mutilateConfig)
			achievedLoad, achievedSLI, err := mutilateLauncher.Tune(5000) // very high to be easile achiveable
			So(err, ShouldBeNil)
			So(achievedLoad, ShouldNotEqual, 0)
			So(achievedSLI, ShouldNotEqual, 0)
			_, currentGetCount := getMemcachedStats(mcAddress, t)
			So(currentGetCount, ShouldBeGreaterThan, previousGetCnt)
		})

		Convey("When run mutilate load", func() {
			_, previousGetCnt := getMemcachedStats(mcAddress, t)
			mutilateLauncher := mutilate.New(executor.NewLocal(), mutilateConfig)
			mutilateHandle, err := mutilateLauncher.Load(10, 1*time.Second)
			So(err, ShouldBeNil)
			mutilateHandle.Wait(0)
			out, err := mutilateHandle.StdoutFile()
			log.Println(out.Name())
			rawMetrics, err := parse.File(out.Name())

			SoNonZeroMetricExists := func(name string) {
				v, ok := rawMetrics[name]
				So(ok, ShouldBeTrue)
				So(v, ShouldBeGreaterThan, 0)
			}

			SoNonZeroMetricExists("qps")
			SoNonZeroMetricExists("avg")
			SoNonZeroMetricExists("std")
			SoNonZeroMetricExists("min")
			SoNonZeroMetricExists("percentile/99th")
			SoNonZeroMetricExists("percentile/99.000th/custom")

			if err := mutilateHandle.EraseOutput(); err != nil {
				t.Fatal(err)

			}
			_, currentGetCount := getMemcachedStats(mcAddress, t)
			So(currentGetCount, ShouldBeGreaterThan, previousGetCnt)
			if exitcode, err := mutilateHandle.ExitCode(); err != nil || exitcode != 0 {
				t.Fatalf("mutilate didn't stopped correclty err=%q, exitcode=%d", err, exitcode)
			}
		})
	})
}

// getMemcachedStats helper read and parse "stats" memcached command and return map key -> value
// https://github.com/memcached/memcached/blob/master/doc/protocol.txt#L511
func getMemcachedStats(mcAddress string, t *testing.T) (currItems, getCount int) {

	const (
		statsCmd         = "stats\n"
		mcStatsReplySize = 4096 // enough to get whole response from memcache
	)

	conn, err := net.Dial("tcp", mcAddress)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if n, err := conn.Write([]byte(statsCmd)); err != nil || n != len(statsCmd) {
		t.Fatalf("coulnd't write to memcached exepected number err=%s of bytes=%d", err, n)
	}

	buf := make([]byte, mcStatsReplySize)
	if _, err = conn.Read(buf); err != nil {
		t.Fatal(err)
	}

	for _, line := range strings.Split(string(buf), "\n") {
		if strings.HasPrefix(line, "END") {
			break
		}
		var key, value string
		_, err := fmt.Sscanf(line, "STAT %s %s", &key, &value)
		if err != nil {
			t.Fatal(err)
		}
		switch key {
		case "curr_items":
			if currItems, err = strconv.Atoi(value); err != nil {
				t.Fatal(err)
			}
		case "cmd_get":
			if getCount, err = strconv.Atoi(value); err != nil {
				t.Fatal(err)
			}
		}
	}

	return
}
