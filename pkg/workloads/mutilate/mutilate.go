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
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	"github.com/intelsdi-x/swan/pkg/utils/os"
	"github.com/intelsdi-x/swan/pkg/workloads"
	"github.com/shopspring/decimal"
	"path"
)

const (
	defaultMemcachedHost          = "127.0.0.1"
	defaultMemcachedPercentile    = "99.9"
	defaultMemcachedTuningTimeSec = 10
	defaultMutilatePath           = "data_caching/memcached/mutilate/mutilate"
	mutilatePathEnv               = "SWAN_MUTILATE_PATH"
)

// GetPathFromEnvOrDefault returns the mutilate binary path from environment variable
// SWAN_MUTILATE_PATH or default path in swan directory.
func GetPathFromEnvOrDefault() string {
	return os.GetEnvOrDefault(
		mutilatePathEnv, path.Join(fs.GetSwanWorkloadsPath(), defaultMutilatePath))
}

// Config contains all data for running mutilate.
type Config struct {
	MutilatePath      string
	MemcachedHost     string
	TuningTime        time.Duration
	LatencyPercentile decimal.Decimal
}

// DefaultMutilateConfig is a constructor for MutilateConfig with default parameters.
func DefaultMutilateConfig() Config {
	percentile, _ := decimal.NewFromString(defaultMemcachedPercentile)
	return Config{
		MutilatePath:      GetPathFromEnvOrDefault(),
		MemcachedHost:     defaultMemcachedHost,
		LatencyPercentile: percentile,
		TuningTime:        defaultMemcachedTuningTimeSec * time.Second,
	}
}

type mutilate struct {
	executor executor.Executor
	config   Config
}

// New returns a new Mutilate Load Generator instance.
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

	exitCode, err := taskHandle.ExitCode()
	if err != nil {
		return err
	}

	if exitCode != 0 {
		return errors.New("Memchaced population exited with code: " +
			strconv.Itoa(exitCode))
	}

	return err
}

// Tune returns the maximum achieved QPS where SLI is below target SLO.
func (m mutilate) Tune(slo int) (qps int, achievedSLI int, err error) {
	tuneCmd := m.getTuneCommand(slo)

	fmt.Println(tuneCmd)
	taskHandle, err := m.executor.Execute(tuneCmd)
	if err != nil {
		errMsg := fmt.Sprintf("Executing Mutilate Tune command %s failed; ", tuneCmd)
		return qps, achievedSLI, errors.New(errMsg + err.Error())
	}

	taskHandle.Wait(0)

	exitCode, err := taskHandle.ExitCode()
	if err != nil {
		return qps, achievedSLI, err
	}

	if exitCode != 0 {
		return qps, achievedSLI, errors.New(
			"Executing Mutilate Tune command returned with exit code: " +
				strconv.Itoa(exitCode))
	}

	stdoutFile, err := taskHandle.StdoutFile()
	if err != nil {
		return qps, achievedSLI, err
	}
	qps, achievedSLI, err = getQPSAndLatencyFrom(stdoutFile)
	if err != nil {
		errMsg := fmt.Sprintf("Could not retrieve QPS from Mutilate Tune output. ")
		return qps, achievedSLI, errors.New(errMsg + err.Error())
	}

	return qps, achievedSLI, err
}

// Load starts a load on the specific workload with the defined loadPoint (number of QPS).
// The task will do the load for specified amount of time.
// Note: Results from Load needs to be fetched out of band e.g using Snap.
func (m mutilate) Load(qps int, duration time.Duration) (executor.TaskHandle, error) {
	return m.executor.Execute(m.getLoadCommand(qps, duration))
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
		return 0, fmt.Errorf("Cannot find regex match in output: %s", output)
	}

	latency, err = strconv.Atoi(match[1])
	if err != nil {
		errMsg := fmt.Sprintf(
			"Cannot parse integer from string: %s; ", match[1])
		return 0, errors.New(errMsg + err.Error())
	}

	return latency, err
}
