package mutilate

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads"
	"github.com/shopspring/decimal"
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

	exitCode, err := taskHandle.GetExitCode()
	if err != nil {
		return err
	}

	if exitCode != 0 {
		return errors.New("Memchaced population exited with code: " +
			strconv.Itoa(exitCode))
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

	exitCode, err := taskHandle.GetExitCode()
	if err != nil {
		return qps, achievedSLI, err
	}

	if exitCode != 0 {
		return qps, achievedSLI, errors.New(
			"Executing Mutilate Tune command returned with exit code: " +
				strconv.Itoa(exitCode))
	}

	stdoutReader, err := taskHandle.Stdout()
	if err != nil {
		return qps, achievedSLI, err
	}
	qps, achievedSLI, err = getQPSAndLatencyFrom(stdoutReader)
	if err != nil {
		errMsg := fmt.Sprintf("Could not retrieve QPS from Mutilate Tune output. ")
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

	exitCode, err := taskHandle.GetExitCode()
	if err != nil {
		return achievedQPS, sli, err
	}

	if exitCode != 0 {
		errMsg := fmt.Sprintf("Executing Mutilate Load returned with exit code %d", exitCode)
		return achievedQPS, sli, errors.New(errMsg + err.Error())
	}

	stdoutReader, err := taskHandle.Stdout()
	if err != nil {
		return achievedQPS, sli, err
	}

	achievedQPS, sli, err = getQPSAndLatencyFrom(stdoutReader)
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

func getQPSAndLatencyFrom(outputReader io.Reader) (qps int, latency int, err error) {

	buff, err := ioutil.ReadAll(outputReader)
	if err != nil {
		return qps, latency, err
	}
	output := string(buff)
	qps, qpsError := getQPSFrom(output)
	latency, latencyError := getLatencyFrom(output)

	var errorMsg string

	if qpsError != nil {
		errorMsg += "Could not get QPS from output: " + qpsError.Error() + ". "
	}
	if latencyError != nil {
		errorMsg += "Could not get Latency from output: " + latencyError.Error() + ". "
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
