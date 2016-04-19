package workloads

import (
	"errors"
	"fmt"
	"github.com/intelsdi-x/swan/pkg/executor"
	"regexp"
	"strconv"
	"time"
)

/**
Real deployments of memcached often handle the requests of dozens, hundreds, or thousands of front-end clients simultaneously. However, by default, mutilate establishes one connection per server and meters requests one at a time (it waits for a reply before sending the next request). This artificially limits throughput (i.e. queries per second), as the round-trip network latency is almost certainly far longer than the time it takes for the memcached server to process one request.

In order to get reasonable benchmark results with mutilate, it needs to be configured to more accurately portray a realistic client workload. In general, this means ensuring that (1) there are a large number of client connections, (2) there is the potential for a large number of outstanding requests, and (3) the memcached server saturates and experiences queuing delay far before mutilate does. I suggest the following guidelines:

    Establish on the order of 100 connections per memcached server thread.
    Don't exceed more than about 16 connections per mutilate thread.
    Use multiple mutilate agents in order to achieve (1) and (2).
    Do not use more mutilate threads than hardware cores/threads.
    Use -Q to configure the "master" agent to take latency samples at slow, a constant rate.

https://github.com/leverich/mutilate

*/

type MutilateConfig struct {
	mutilate_path      string
	memcached_uri      string
	tuning_time        time.Duration
	latency_percentile int
}

func DefaultMutilateConfig() MutilateConfig {
	return MutilateConfig{
		mutilate_path:      "mutilate",
		memcached_uri:      "localhost",
		tuning_time:        10 * time.Second,
		latency_percentile: 999,
	}
}

type Mutilate struct {
	executor executor.Executor
	config   MutilateConfig
}

func NewMutilate(executor executor.Executor, config MutilateConfig) Mutilate {
	return Mutilate{
		executor: executor,
		config:   config,
	}
}

func (m *Mutilate) Populate() error {
	popCmd := fmt.Sprintf("mutilate -s %s --loadonly",
		m.config.memcached_uri)
	taskHandle, err := m.executor.Execute(popCmd)
	if err != nil {
		return err
	}
	taskHandle.Wait(0)
	// TODO(skonefal): Check exit code

	return nil
}

func (m Mutilate) Tune(slo int) (targetQPS int, err error) {
	// mutilate -s localhost --search=999:1000
	tuneCmd := fmt.Sprintf("%s -s %s --search=%d:%d -t %d",
		m.config.mutilate_path,
		m.config.memcached_uri,
		m.config.latency_percentile,
		slo,
		int(m.config.tuning_time.Seconds()))

	taskHandle, err := m.executor.Execute(tuneCmd)
	if err != nil {
		errMsg := fmt.Sprintf(
			"Executing Mutilate Tune cmd %s failed; ", tuneCmd)
		return targetQPS, errors.New(errMsg + err.Error())
	}
	taskHandle.Wait(0)

	_, status := taskHandle.Status()
	if status == nil {
		panic("something wrong, debug")
	}

	output := status.Stdout
	targetQPS, err = getTargetQps(output)
	if err != nil {

	}

	return targetQPS, err
}

func getTargetQps(mutilateOutput string) (targetQps int, err error) {
	//Total QPS = 4450.3 (89007 / 20.0s)
	re := regexp.MustCompile(`Total QPS =\s(\d+)`)
	match := re.FindStringSubmatch(mutilateOutput)

	if matchNotFound(match) {
		errMsg := fmt.Sprintf(
			"Cannot find targetQPS in output: %s", mutilateOutput)
		return targetQps, errors.New(errMsg)
	}

	targetQps, err = strconv.Atoi(match[1])
	if err != nil {
		errMsg := fmt.Sprintf(
			"Cannot parse targetQPS in string: %s; ", match[1])
		return targetQps, errors.New(errMsg + err.Error())
	}

	return targetQps, err
}

func matchNotFound(match []string) bool {
	return match == nil || len(match[1]) == 0
}

func (m Mutilate) Load(qps int, duration time.Duration) (sli int, err error) {
	command := fmt.Sprintf("mutilate -m")
	taskHandle, error := m.executor.Execute(command)

	_ = taskHandle
	_ = error
	return -1, nil
}

func getLatencyFromMutilateOutputLine(line string) (latencies []int) {
	//re := regexp.MustCompile(`\d+.\d+\s(\d+)`)
	return latencies
}

func getLatenciesFromMutilateOutput(input string) (latencies []int) {

	return latencies
}

func computeHistogramLatency(percentile int, input string) (latency int) {
	_ = input
	return latency
}
