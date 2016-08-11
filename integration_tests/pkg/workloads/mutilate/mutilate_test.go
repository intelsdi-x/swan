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
	"github.com/intelsdi-x/swan/pkg/utils/env"
	"github.com/intelsdi-x/swan/pkg/utils/err_collection"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	memachedPort = 11212
)

// TestMutilateWithExecutor is an integration test with local executor.
// Flow:
// - start memcached and make sure it is a new clean instance
// - run populate and check new items were stored
// - run tune - search for capacity and ensure it is not zero
// - run load and check it run without error (ignore results)
// note: for Populate/Tune we don't check output files.
func TestMutilateWithExecutor(t *testing.T) {
	//log.SetLevel(log.DebugLevel)
	//log.SetOutput(os.Stderr)

	// Memcached setup.
	memcachedConfig := memcached.DefaultMemcachedConfig()
	memcachedConfig.User = env.GetOrDefault("USER", memcachedConfig.User)
	memcachedConfig.Port = memachedPort
	mcAddress := fmt.Sprintf("127.0.0.1:%d", memcachedConfig.Port)

	// Start memcached and make sure it is a new one.
	memcachedLauncher := memcached.New(executor.NewLocal(), memcachedConfig)
	mcHandle, err := memcachedLauncher.Launch()

	// Clean after memcached ...
	defer func() {
		// ... prevent before stopping, cleaning up and erasing output from empty task handle ...
		So(mcHandle, ShouldNotBeNil)

		var errCollection errcollection.ErrorCollection
		// and our memcached instance was closed properly.
		errCollection.Add(mcHandle.Stop())
		mcHandle.Wait(0)

		ec, err := mcHandle.ExitCode()
		errCollection.Add(err)
		if ec != -1 {
			// Expect -1 on SIGKILL (TODO: change to zero, after Stop "graceful stop fix").
			t.Fatalf("memcached was stopped incorrectly err %s exit-code: %d", err, ec)
		}

		// Make sure that fd are closed
		errCollection.Add(mcHandle.Clean())
		// Make sure temp files removal was successful.
		errCollection.Add(mcHandle.EraseOutput())

		So(errCollection.GetErrIfAny(), ShouldBeNil)
	}()

	if err != nil {
		t.Fatal("cannot start memcached:" + err.Error())
	}

	// Give memcached chance to start and possibly die.
	if stopped := mcHandle.Wait(1 * time.Second); stopped {
		t.Fatal("memcached is not running after the second")
	}

	currItems, _ := getMemcachedStats(mcAddress, t)
	if currItems != 0 { // In case of not empty or someone at the same time is messing with it.
		t.Fatal("expecting empty memcached but some items are already there")
	}

	Convey("With memcached started already", t, func() {

		mutilateConfig := mutilate.DefaultMutilateConfig()
		mutilateConfig.TuningTime = 1 * time.Second
		// Ensure files are removed afterwards.
		mutilateConfig.ErasePopulateOutput = true
		mutilateConfig.EraseTuneOutput = true
		mutilateConfig.WarmupTime = 1 * time.Second
		// Note: not sure if custom percentile is working correctly.
		// TODO: added a custom percentile integration test.
		mutilateConfig.LatencyPercentile = "99.1234"
		mutilateConfig.MemcachedPort = memcachedConfig.Port
		mutilateConfig.ErasePopulateOutput = true
		mutilateConfig.EraseTuneOutput = true

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
			// Tune up to 5000 us to be be easily achievable.
			achievedLoad, achievedSLI, err := mutilateLauncher.Tune(5000)
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
				v, ok := rawMetrics.Raw[name]
				So(ok, ShouldBeTrue)
				So(v, ShouldBeGreaterThan, 0)
			}

			SoNonZeroMetricExists("qps")
			SoNonZeroMetricExists("avg")
			SoNonZeroMetricExists("std")
			SoNonZeroMetricExists("min")
			SoNonZeroMetricExists("percentile/99th")
			SoNonZeroMetricExists("percentile/custom")

			if err := mutilateHandle.EraseOutput(); err != nil {
				t.Fatal(err)

			}
			_, currentGetCount := getMemcachedStats(mcAddress, t)
			So(currentGetCount, ShouldBeGreaterThan, previousGetCnt)
			if exitcode, err := mutilateHandle.ExitCode(); err != nil || exitcode != 0 {
				t.Fatalf("mutilate didn't stopped correctly err=%q, exitcode=%d", err, exitcode)
			}
		})
	})
}

// getMemcachedStats helper read and parse "stats" memcached command and return map key -> value.
// https://github.com/memcached/memcached/blob/master/doc/protocol.txt#L511
func getMemcachedStats(mcAddress string, t *testing.T) (currItems, getCount int) {
	const (
		statsCmd         = "stats\n"
		mcStatsReplySize = 4096 // Enough size to get whole response from memcached.
	)

	conn, err := net.Dial("tcp", mcAddress)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if n, err := conn.Write([]byte(statsCmd)); err != nil || n != len(statsCmd) {
		t.Fatalf("couldn't write to memcached expected number err=%s of bytes=%d", err, n)
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
