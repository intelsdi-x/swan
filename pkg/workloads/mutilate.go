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

// MutilateConfig contains all data for running mutilate.
type MutilateConfig struct {
	mutilatePath      string
	memcachedHost     string
	tuningTime        time.Duration
	latencyPercentile float64
}

// DefaultMutilateConfig ...
func DefaultMutilateConfig() MutilateConfig {
	return MutilateConfig{
		mutilatePath:      "mutilate",
		memcachedHost:     "localhost",
		tuningTime:        10 * time.Second,
		latencyPercentile: 99.9,
	}
}

// Mutilate is a load generator for Memcached.
// https://github.com/leverich/mutilate
type Mutilate struct {
	executor executor.Executor
	config   MutilateConfig
}

// NewMutilate ...
func NewMutilate(executor executor.Executor, config MutilateConfig) Mutilate {
	return Mutilate{
		executor: executor,
		config:   config,
	}
}

// Populate loads Memcached and exits.
func (m *Mutilate) Populate() error {
	popCmd := fmt.Sprintf("%s -s %s --loadonly",
		m.config.mutilatePath,
		m.config.memcachedHost,
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

// Tune returns maximum QPSes for given SLO
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

// Load exercises `qps` load on Memcached for given duration.
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
		m.config.latencyPercentile, 'f', -1, 64)
	return fmt.Sprintf("%s -s %s -q %d -t %d --swanpercentile=%s",
		m.config.mutilatePath,
		m.config.memcachedHost,
		qps,
		int(duration.Seconds()),
		percentile,
	)
}

func (m Mutilate) getTuneCommand(slo int) (command string, err error) {
	mutilateSearchPercentile, err := transformFloatToIntegerWithoutDot(
		m.config.latencyPercentile)
	if err != nil {
		return command, errors.New("Parse percentile value failed")
	}
	command = fmt.Sprintf("%s -s %s --search=%d:%d -t %d",
		m.config.mutilatePath,
		m.config.memcachedHost,
		mutilateSearchPercentile,
		slo,
		int(m.config.tuningTime.Seconds()),
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
