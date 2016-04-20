package workloads

import (
	"errors"
	"fmt"
	"github.com/intelsdi-x/swan/pkg/executor"
	"regexp"
	"strconv"
	"strings"
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
	latency_percentile float64
}

func DefaultMutilateConfig() MutilateConfig {
	return MutilateConfig{
		mutilate_path:      "mutilate",
		memcached_uri:      "localhost",
		tuning_time:        10 * time.Second,
		latency_percentile: 99.9,
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
	popCmd := fmt.Sprintf("%s -s %s --loadonly",
		m.config.mutilate_path,
		m.config.memcached_uri,
	)
	taskHandle, err := m.executor.Execute(popCmd)
	if err != nil {
		return err
	}
	taskHandle.Wait(0)

	_, status := taskHandle.Status()
	if status.ExitCode != 0 {
		errMsg := fmt.Sprintf("Memchaced populate exited with code: %d", status.ExitCode)
		return errors.New(errMsg)
	}

	return nil
}

func (m Mutilate) Tune(slo int) (targetQPS int, err error) {
	tuneCmd, err := m.getTuneCommand(slo)
	if err != nil {
		errMsg := fmt.Sprintf("Preparation of Tune command failed: %s", err)
		return targetQPS, errors.New(errMsg)
	}

	taskHandle, err := m.executor.Execute(tuneCmd)
	if err != nil {
		errMsg := fmt.Sprintf("Executing Mutilate Tune command %s failed; ", tuneCmd)
		return targetQPS, errors.New(errMsg + err.Error())
	}

	taskHandle.Wait(0)

	_, status := taskHandle.Status()
	if status.ExitCode != 0 {
		errMsg := fmt.Sprintf(
			"Executing Mutilate Tune command returned with exit code %d", status.ExitCode)
		return targetQPS, errors.New(errMsg + err.Error())
	}

	re := regexp.MustCompile(`Total QPS =\s(\d+)`)
	targetQPS, err = getValueFromOutput(status.Stdout, re)
	if err != nil {
		errMsg := fmt.Sprintf("Could not retrieve QPS from Mutilate Tune output")
		return targetQPS, errors.New(errMsg + err.Error())
	}

	return targetQPS, err
}

func (m Mutilate) Load(qps int, duration time.Duration) (sli int, err error) {
	loadCmd := m.getLoadCommand(qps, duration)

	taskHandle, err := m.executor.Execute(loadCmd)
	if err != nil {
		errMsg := fmt.Sprintf("Executing Mutilate Tune command %s failed; ", loadCmd)
		return sli, errors.New(errMsg + err.Error())
	}

	taskHandle.Wait(0)

	_, status := taskHandle.Status()
	if status.ExitCode != 0 {
		errMsg := fmt.Sprintf("Executing Mutilate Load returned with exit code %d",
			status.ExitCode)
		return sli, errors.New(errMsg + err.Error())
	}

	re := regexp.MustCompile(`Swan latency for percentile \d+.\d+:\s(\d+)`)
	sli, err = getValueFromOutput(status.Stdout, re)
	if err != nil {
		errMsg := fmt.Sprintf("Could not retrieve SLI from mutilate Load output")
		return sli, errors.New(errMsg + err.Error())
	}

	return sli, err
}

// Mutilate Search method requires percentile in an integer format
// Transforms (float)999.9 to (int)9999
func transformFloatToIntegerWithoutDot(value float64) (ret int64, err error) {
	percentileString := strconv.FormatFloat(value, 'f', -1, 64)
	percentileWithoutDot := strings.Replace(percentileString, ".", "", -1)
	tunePercentile, err := strconv.ParseInt(percentileWithoutDot, 10, 64)
	if err != nil {
		return 0, errors.New("Parse value failed")
	}

	return tunePercentile, err
}

func (m Mutilate) getLoadCommand(qps int, duration time.Duration) string {
	percentile := strconv.FormatFloat(
		m.config.latency_percentile, 'f', -1, 64)
	return fmt.Sprintf("%s -s %s -q %d -t %d --swanpercentile=%f",
		m.config.mutilate_path,
		m.config.memcached_uri,
		qps,
		duration.Seconds(),
		percentile,
	)
}

func (m Mutilate) getTuneCommand(slo int) (command string, err error) {
	mutilateSearchPercentile, err := transformFloatToIntegerWithoutDot(
		m.config.latency_percentile)
	if err != nil {
		return command, errors.New("Parse percentile value failed")
	}
	command = fmt.Sprintf("%s -s %s --search=%d:%d -t %d",
		m.config.mutilate_path,
		m.config.memcached_uri,
		mutilateSearchPercentile,
		slo,
		int(m.config.tuning_time.Seconds()),
	)
	return command, err
}

func matchNotFound(match []string) bool {
	return match == nil || len(match[1]) == 0
}

func getValueFromOutput(output string, regex *regexp.Regexp) (value int, err error) {
	match := regex.FindStringSubmatch(output)
	if matchNotFound(match) {
		errMsg := fmt.Sprintf(
			"Cannot find regex in output: %s", output)
		return value, errors.New(errMsg)
	}

	value, err = strconv.Atoi(match[1])
	if err != nil {
		errMsg := fmt.Sprintf(
			"Cannot parse value in string: %s; ", match[1])
		return value, errors.New(errMsg + err.Error())
	}

	return value, err
}
