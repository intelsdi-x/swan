package mutilate

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	snapmutilate "github.com/intelsdi-x/swan/misc/snap-plugin-collector-mutilate/mutilate"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/utils/os"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	"github.com/shopspring/decimal"
	. "github.com/smartystreets/goconvey/convey"
)

// getMemcachedStats helper read and parse "stats" memcached command and return map key -> value
// https://github.com/memcached/memcached/blob/master/doc/protocol.txt#L511
func getMemcachedStats(mcAddress string, t *testing.T) (currItems, getCount int) {

	conn, err := net.Dial("tcp", mcAddress)
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()
	n, err := conn.Write([]byte("stats\n"))
	if err != nil {
		t.Error(err)
	}
	if n != 6 {
		t.Error("coulnd't write to memcache exepected number of bytes:" + strconv.Itoa(n))
	}

	buf := make([]byte, 2048)
	n, err = conn.Read(buf)
	if err != nil {
		t.Error(err)
	}
	stats := make(map[string]string)
	for _, line := range strings.Split(string(buf), "\n") {
		fields := strings.Fields(line)
		// stats dump ends with END keyword - it means we're done
		if len(fields) == 1 && fields[0] == "END" {
			break
		}
		key, value := fields[1], fields[2]
		stats[key] = value
	}

	currItems, err = strconv.Atoi(stats["curr_items"])
	if err != nil {
		t.Error(err)
	}

	getCount, err = strconv.Atoi(stats["cmd_get"])
	if err != nil {
		t.Error(err)
	}
	return
}

// TestMutilateWithExecutor is an integration test with local executor.
// Flow:
// - start memcache and make sure is new clean instance
// - run populate and check new items were stored
// - run tune - search for capacity and make it is not zero
// - run load and check it run without error (ignore results)
// note: couldn't check output files because cannot inject check coulnd before removal
func TestMutilateWithExecutor(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	// memcached setup
	mcConfig := memcached.DefaultMemcachedConfig()
	mcConfig.User = os.GetEnvOrDefault("USER", mcConfig.User)
	mcAddress := fmt.Sprintf("127.0.0.1:%d", mcConfig.Port)
	// closure function with memcache address for getting statistics from mecache

	// start memcached and make sure it is a new one!
	memcachedLauncher := memcached.New(executor.NewLocal(), mcConfig)
	mcHandle, err := memcachedLauncher.Launch()
	if err != nil {
		t.Error("cannot start memcached:" + err.Error())
	}

	errorOnError := func(f func() error) {
		if err := f(); err != nil {
			t.Error(err)
		}
	}
	defer func() {
		errorOnError(mcHandle.EraseOutput) // make sure temp files removal was successful
		errorOnError(mcHandle.Stop)        // and our memcached instance was closed properlly
		mcHandle.Wait(0)
		ec, err := mcHandle.ExitCode()
		if err != nil {
			t.Error(err)
		}
		if ec != -1 { // expected -1 on SIGKILL (TODO: change to zero, after Stop "gracefull stop fix"
			t.Error(fmt.Sprintf("memcached was stopped incorrectly exit-code: %d", ec))
		}
	}()

	// give memcache chance to start and possibly die
	if stopped := mcHandle.Wait(1 * time.Second); stopped {
		t.Error("my memcached is not running after the second")
	}

	if mcHandle.Status() != executor.RUNNING {
		t.Error("my memcached is not running!")
	}
	currItems, _ := getMemcachedStats(mcAddress, t)
	if currItems != 0 { // in case of not my memache or someone at the same time is messing with it
		t.Error("expecting empty mc! but some items are already there")
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

		_, previousGetCnt := getMemcachedStats(mcAddress, t)
		Convey("When run mutilate tune", func() {
			mutilateLauncher := mutilate.New(executor.NewLocal(), mutilateConfig)
			achievedLoad, achievedSLI, err := mutilateLauncher.Tune(5000) // very high to be easile achiveable
			So(err, ShouldBeNil)
			So(achievedLoad, ShouldNotEqual, 0)
			So(achievedSLI, ShouldNotEqual, 0)
			_, currentGetCount := getMemcachedStats(mcAddress, t)
			So(currentGetCount, ShouldBeGreaterThan, previousGetCnt)
		})

		_, previousGetCnt = getMemcachedStats(mcAddress, t)
		Convey("When run mutilate load", func() {
			mutilateLauncher := mutilate.New(executor.NewLocal(), mutilateConfig)
			mutilateHandle, err := mutilateLauncher.Load(10, 1*time.Second)
			So(err, ShouldBeNil)
			mutilateHandle.Wait(0)
			out, err := mutilateHandle.StdoutFile()
			log.Println(out.Name())
			rawMetrics, err := snapmutilate.ParseOutput(out.Name())

			for _, expectedMetric := range []string{"qps", "avg", "std", "min", "percentile/99th", "percentile/99.000th/custom"} {
				v, ok := rawMetrics[expectedMetric]
				So(ok, ShouldBeTrue)
				So(v, ShouldBeGreaterThan, 0)
				log.Debugf("%s = %0.2f", expectedMetric, v)
			}

			errorOnError(mutilateHandle.EraseOutput)
			_, currentGetCount := getMemcachedStats(mcAddress, t)
			So(currentGetCount, ShouldBeGreaterThan, previousGetCnt)
			if exitcode, err := mutilateHandle.ExitCode(); err != nil || exitcode != 0 {
				t.Fatalf("mutilate didn't stopped correclty err=%q, exitcode=%d", err, exitcode)
			}
		})
	})
}
