package main

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/snap"
	"github.com/intelsdi-x/athena/pkg/utils/errutil"
	"github.com/intelsdi-x/swan/experiments/memcached-sensitivity-profile/common"
	"github.com/intelsdi-x/swan/experiments/memcached-sensitivity-profile/topology"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/montanaflynn/stats"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/smartystreets/goconvey/convey"
)

func Seq(desc string, from, to, step int, f func(i int)) {
	for i := 0; i < to; i = i + step {
		Convey(desc+" i = "+strconv.Itoa(i), func() {
			f(i)
		})
	}
}

func Range(desc string, iter []interface{}, f func(i interface{})) {
	for _, i := range iter {
		Convey(desc+" i = "+fmt.Sprintf("%s", i), func() {
			f(i)
		})
	}
}

const name = "memcached-sensitivity-profile"

func main() {

	// ------------------- Bootstrap -------------------------
	conf.SetAppName(name)
	err := conf.ParseFlags()
	errutil.Check(err)

	log.SetLevel(conf.LogLevel())
	// validate.OS()

	// ------------------- Isolations configuration ----------------
	hpIsolation, l1Isolation, llcIsolation := topology.NewIsolations()

	// ------------------- Executors configuration ----------------
	hpExecutor, beExecutorFactory, cleanup, err := common.PrepareExecutors(hpIsolation)
	errutil.Check(err)
	defer cleanup()

	// ------------------ Aggressors configuration ---------------
	aggressorSessionLaunchersPairs, err := common.PrepareAggressors(l1Isolation, llcIsolation, beExecutorFactory)
	errutil.Check(err)

	// ------------------ HP workload configuration ---------------------------
	memcachedConfig := memcached.DefaultMemcachedConfig()
	hpLauncher := memcached.New(hpExecutor, memcachedConfig)

	// ------------------ Load generator configuration --------------------------
	loadGenerator, err := common.PrepareMutilateGenerator(memcachedConfig.IP, memcachedConfig.Port)
	errutil.Check(err)
	loadGeneratorSessionLauncher, err := common.PrepareSnapMutilateSessionLauncher()
	errutil.Check(err)

	// ------------------ UUID --------------------------------
	uuid, err := uuid.NewV4()
	errutil.Check(err)
	fmt.Println(uuid)

	// tmp folder (whole experiment resopsbility)
	wd, _ := os.Getwd()
	experimentDirectory := path.Join("/tmp", name, uuid.String())
	errutil.Check(os.MkdirAll(experimentDirectory, 0777))
	errutil.Check(os.Chdir(experimentDirectory))
	defer os.Chdir(wd)

	// Tunning Phase
	peakload := tunning(hpLauncher, loadGenerator)

	// ------------------ declarative  -------------------------
	t := &testing.T{} // just a hack to force running convey in main.
	Convey("experiment ", t, func() {
		Range("aggressor:", aggressorSessionLaunchersPairs, func(aggr interface{}) {
			Seq("qps", 1000, peakload, 1000, func(qps int) {
				Seq("rep", 1, 3, 1, func(rep int) {
					// Measrument Phase
					measurment(rep, qps, aggr, hpLauncher, loadGenerator, loadGeneratorSessionLauncher)
				})
			})
		})
	})

	return

	// ---------------------------------- ALTERNATIVES --------------------------------
	// ----------------- impeartive style --------------------------
	for rep := 0; rep < 3; rep++ {
		for qps := 1000; qps < peakload; qps += 1000 {
			for _, aggr := range aggressorSessionLaunchersPairs {
				measurment(rep, qps, aggr, hpLauncher, loadGenerator, loadGeneratorSessionLauncher)
			}
		}
	}

	nop := func(_ ...interface{}) interface{} { return nil }
	apply, seq, unfold := nop, nop, nop

	// ----------------- functional style --------------------------
	apply(
		unfold(
			seq(0, 3, 1),                   // rep
			seq(1000, peakload, 1000),      // qps
			aggressorSessionLaunchersPairs, // aggr
		),
		func(rep, qps int, aggr interface{}) {
			measurment(rep, qps, aggr, hpLauncher, loadGenerator, loadGeneratorSessionLauncher)
		},
	)

}

// return peakload
func tunning(hpLauncher executor.Launcher, loadGenerator executor.LoadGenerator) int {
	SLO := 500
	loads := []float64{}
	hpTask, err := hpLauncher.Launch()
	errutil.Check(err)

	err = loadGenerator.Populate()
	errutil.Check(err)

	load, _, err := loadGenerator.Tune(SLO)
	errutil.Check(err)
	executor.StopCleanAndErase(hpTask)

	loads = append(loads, float64(load))

	peakload, err := stats.Mean(loads)
	errutil.Check(err)
	log.Debug("Calculated targetLoad (PeakLoadSatisfyingSLO) (QPS/RPS): ", peakload, " for SLO: ", SLO)
	return int(peakload)
}

func measurment(rep, qps int, aggr interface{}, hpLauncher executor.Launcher, loadGenerator executor.LoadGenerator, loadGeneratorSessionLauncher snap.SessionLauncher) {
	tags := "rep=" + strconv.Itoa(rep) + ",qps=" + strconv.Itoa(qps)
	log.Info(tags, aggr)

	// ----------------------------- HP ---------------------------
	prTask, err := hpLauncher.Launch()
	errutil.Check(err)
	defer executor.StopCleanAndErase(prTask)

	// ----------------------------- Aggressors ---------------------------
	aggrPair, ok := aggr.(sensitivity.LauncherSessionPair)
	if ok { // handle baseline
		aggrTaskHandle, err := aggrPair.Launcher.Launch()
		errutil.Check(err)
		defer executor.StopCleanAndErase(aggrTaskHandle)
	}
	//
	// ---------------------------- populate --------------------
	errutil.Check(loadGenerator.Populate())

	// ---------------------------- load generation ----------------------
	loadGeneratorTask, err := loadGenerator.Load(qps, 15*time.Second)
	errutil.Check(err)

	// Wait for load generation to end.
	loadGeneratorTask.Wait(0)
	defer executor.StopCleanAndErase(loadGeneratorTask)

	// // Launch Snap Session for loadGenerator if specified.
	// // NOTE: Common loadGenerators don't have HTTP and just save output to the file after
	// // completing the load generation.
	// // To have our snap task not disabled by Snap daemon because we could not read the file during
	// // load, we need to run snap session (task) only after load generation work ended.
	lgSessionHandle, err := loadGeneratorSessionLauncher.LaunchSession(loadGeneratorTask, tags)
	errutil.Check(err)
	lgSessionHandle.Stop()
}
