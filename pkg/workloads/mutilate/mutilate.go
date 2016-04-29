package mutilate

import (
	"errors"
	"fmt"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads"
	"github.com/shopspring/decimal"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Config contains all data for running mutilate.
type Config struct {
	MutilatePath      string
	MemcachedHost     string
	TuningTime        time.Duration
	LatencyPercentile decimal.Decimal
}

type mutilate struct {
	executor executor.Executor
	config   Config
}

// New returns a new Mutilate Load Generator instance
// Mutilate is a load generator for Memcached.
// https://github.com/leverich/mutilate
func New(executor executor.Executor, config Config) workloads.LoadGenerator {
	return mutilate{
		executor: executor,
		config:   config,
	}
}

// Populate loads Memcached and exits.
func (m mutilate) Populate() (err error) {
	populateCmd := fmt.Sprintf("%s -s %s --loadonly",
		m.config.MutilatePath,
		m.config.MemcachedHost,
	)

	taskHandle, err := m.executor.Execute(populateCmd)
	if err != nil {
		return err
	}
	taskHandle.Wait(0)

	_, status := taskHandle.Status()
	if status.ExitCode != 0 {
		return errors.New("Memchaced population exited with code: " +
			strconv.Itoa(status.ExitCode))
	}

	return err
}

// Tune returns the maximum achieved QPS where SLI is below target SLO
func (m mutilate) Tune(slo int) (qps int, achievedSLI int, err error) {
	tuneCmd := m.getTuneCommand(slo)

	taskHandle, err := m.executor.Execute(tuneCmd)
	if err != nil {
		errMsg := fmt.Sprintf("Executing Mutilate Tune command %s failed; ", tuneCmd)
		return qps, achievedSLI, errors.New(errMsg + err.Error())
	}
	taskHandle.Wait(0)

	_, status := taskHandle.Status()
	if status.ExitCode != 0 {
		return qps, achievedSLI, errors.New(
			"Executing Mutilate Tune command returned with exit code: " +
				strconv.Itoa(status.ExitCode))
	}

	qps, achievedSLI, err = getQPSAndLatencyFrom(status.Stdout)
	if err != nil {
		errMsg := fmt.Sprintf("Could not retrieve QPS from Mutilate Tune output")
		return qps, achievedSLI, errors.New(errMsg + err.Error())
	}

	return qps, achievedSLI, err
}

// Load exercises `qps` load on Memcached for given duration.
func (m mutilate) Load(qps int, duration time.Duration) (achievedQPS int, sli int, err error) {
	loadCmd := m.getLoadCommand(qps, duration)

	taskHandle, err := m.executor.Execute(loadCmd)
	if err != nil {
		errMsg := fmt.Sprintf("Mutilate Tuning %s failed; ", loadCmd)
		return achievedQPS, sli, errors.New(errMsg + err.Error())
	}
	defer taskHandle.Clean()

	taskHandle.Wait(0)

	_, status := taskHandle.Status()
	if status.ExitCode != 0 {
		errMsg := fmt.Sprintf("Executing Mutilate Load returned with exit code %d",
			status.ExitCode)
		return achievedQPS, sli, errors.New(errMsg + err.Error())
	}

	achievedQPS, sli, err = getQPSAndLatencyFrom(status.Stdout)
	if err != nil {
		errMsg := fmt.Sprintf("Could not retrieve information from mutilate Load output. ")
		return achievedQPS, sli, errors.New(errMsg + err.Error())
	}

	return achievedQPS, sli, err
}

// Mutilate Search method requires percentile in an integer format
// Transforms (decimal)999.9 to (int)9999
func transformDecimalToIntegerWithoutDot(value decimal.Decimal) (ret int64) {
	percentileString := value.String()
	percentileWithoutDot := strings.Replace(percentileString, ".", "", -1)
	tunePercentile, err := strconv.ParseInt(percentileWithoutDot, 10, 64)
	if err != nil {
		panic("Parsing " + percentileWithoutDot + " failed.")
	}

	return tunePercentile
}

func (m mutilate) getLoadCommand(qps int, duration time.Duration) string {
	return fmt.Sprintf("%s -s %s -q %d -t %d --swanpercentile=%s",
		m.config.MutilatePath,
		m.config.MemcachedHost,
		qps,
		int(duration.Seconds()),
		m.config.LatencyPercentile.String(),
	)
}

func (m mutilate) getTuneCommand(slo int) (command string) {
	// Transforms (decimal)999.9 to (int)9999
	mutilateSearchPercentile :=
		transformDecimalToIntegerWithoutDot(m.config.LatencyPercentile)
	command = fmt.Sprintf("%s -s %s --search=%d:%d -t %d",
		m.config.MutilatePath,
		m.config.MemcachedHost,
		mutilateSearchPercentile,
		slo,
		int(m.config.TuningTime.Seconds()),
	)
	return command
}

func matchNotFound(match []string) bool {
	return match == nil || len(match) < 2 || len(match[1]) == 0
}

func getQPSAndLatencyFrom(output string) (qps int, latency int, err error) {
	qps, qpsError := getQPSFrom(output)
	latency, latencyError := getLatencyFrom(output)

	var errorMsg string

	if qpsError != nil {
		errorMsg += "Could not get QPS from output: " + err.Error() + ". "
	}
	if latencyError != nil {
		errorMsg += "Could not get Latency from output: " + err.Error() + ". "
	}

	if errorMsg != "" {
		return 0, 0, errors.New(errorMsg)
	}

	return qps, latency, err
}

func getQPSFrom(output string) (qps int, err error) {
	getQPSRegex := regexp.MustCompile(`Total QPS =\s(\d+)`)
	match := getQPSRegex.FindStringSubmatch(output)
	if matchNotFound(match) {
		errMsg := fmt.Sprintf(
			"Cannot find regex match in output: %s", output)
		return 0, errors.New(errMsg)
	}

	qps, err = strconv.Atoi(match[1])
	if err != nil {
		errMsg := fmt.Sprintf(
			"Cannot parse integer from string: %s; ", match[1])
		return 0, errors.New(errMsg + err.Error())
	}

	return qps, err
}

func getLatencyFrom(output string) (latency int, err error) {
	getLatencyRegex := regexp.MustCompile(`Swan latency for percentile \d+.\d+:\s(\d+)`)
	match := getLatencyRegex.FindStringSubmatch(output)
	if matchNotFound(match) {
		errMsg := fmt.Sprintf(
			"Cannot find regex match in output: %s", output)
		return 0, errors.New(errMsg)
	}

	latency, err = strconv.Atoi(match[1])
	if err != nil {
		errMsg := fmt.Sprintf(
			"Cannot parse integer from string: %s; ", match[1])
		return 0, errors.New(errMsg + err.Error())
	}

	return latency, err
}
