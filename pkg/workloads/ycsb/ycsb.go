package ycsb

import (
	"fmt"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads/redis"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

const (
	name                               = "YCSB"
	defaultPathToBinary                = "ycsb"
	defaultWorkload                    = "com.yahoo.ycsb.workloads.CoreWorkload"
	defaultWorkloadRecordCount         = 100000
	defaultWorkloadReadAllFields       = true
	defaultWorkloadReadProportion      = "0.5"
	defaultWorkloadUpdateProportion    = "0.5"
	defaultWorkloadScanProportion      = "0.0"
	defaultWorkloadInsertProportion    = "0.0"
	defaultWorkloadRequestDistribution = "zipfian"
)

var (
	pathFlag                        = conf.NewStringFlag("ycsb_path", "Path to YCSB binary file.", defaultPathToBinary)
	workloadFlag                    = conf.NewStringFlag("ycsb_workload", "Name of YCSB workload", defaultWorkload)
	workloadRecordCountFlag         = conf.NewIntFlag("ycsb_workload_recordcount", "Workload record count.", defaultWorkloadRecordCount)
	workloadReadAllFieldsFlag       = conf.NewBoolFlag("ycsb_workload_readallfields", "Workload read all fields.", defaultWorkloadReadAllFields)
	workloadReadProportionFlag      = conf.NewStringFlag("ycsb_workload_readproportion", "Workload read proportion.", defaultWorkloadReadProportion)
	workloadUpdateProportionFlag    = conf.NewStringFlag("ycsb_workload_updateproportion", "Workload update proportion.", defaultWorkloadUpdateProportion)
	workloadScanProportionFlag      = conf.NewStringFlag("ycsb_workload_scanproportion", "Workload scan proportion.", defaultWorkloadScanProportion)
	workloadInsertProportionFlag    = conf.NewStringFlag("ycsb_workload_insertproportion", "Workload insert proportion.", defaultWorkloadInsertProportion)
	workloadRequestDistributionFlag = conf.NewStringFlag("ycsb_workload_requestdistribution", "Workload request distribution.", defaultWorkloadRequestDistribution)
)

type ycsb struct {
	executor executor.Executor
	config   Config
}

// Config is a config for the YCSB.
type Config struct {
	PathToBinary                string
	RedisHost                   string
	RedisPort                   int
	RedisClusterMode            bool // Redis cluster mode.
	Workload                    string
	WorkloadRecordCount         int
	WorkloadOperationCount      int64
	WorkloadReadAllFields       bool
	WorkloadReadProportion      float64
	WorkloadUpdateProportion    float64
	WorkloadScanProportion      float64
	WorkloadInsertProportion    float64
	WorkloadRequestDistribution string
	workloadCommand             string
}

// DefaultYcsbConfig is a constructor for YcsbConfig with default parameters.
func DefaultYcsbConfig() Config {

	workloadReadProportion, err := strconv.ParseFloat(workloadReadProportionFlag.Value(), 64)
	if err != nil {
		workloadReadProportion = 0.0
	}

	workloadUpdateProportion, err := strconv.ParseFloat(workloadUpdateProportionFlag.Value(), 64)
	if err != nil {
		workloadUpdateProportion = 0.0
	}

	workloadScanProportion, err := strconv.ParseFloat(workloadScanProportionFlag.Value(), 64)
	if err != nil {
		workloadScanProportion = 0.0
	}

	workloadInsertProportion, err := strconv.ParseFloat(workloadInsertProportionFlag.Value(), 64)
	if err != nil {
		workloadInsertProportion = 0.0
	}

	return Config{
		PathToBinary:                pathFlag.Value(),
		RedisHost:                   redis.IPFlag.Value(),
		RedisPort:                   redis.PortFlag.Value(),
		RedisClusterMode:            redis.ClusterFlag.Value(),
		Workload:                    workloadFlag.Value(),
		WorkloadRecordCount:         workloadRecordCountFlag.Value(),
		WorkloadReadAllFields:       workloadReadAllFieldsFlag.Value(),
		WorkloadReadProportion:      workloadReadProportion,
		WorkloadUpdateProportion:    workloadUpdateProportion,
		WorkloadScanProportion:      workloadScanProportion,
		WorkloadInsertProportion:    workloadInsertProportion,
		WorkloadRequestDistribution: workloadRequestDistributionFlag.Value(),
	}
}

// New is a contructor for YCSB.
func New(exec executor.Executor, config Config) executor.LoadGenerator {
	return ycsb{
		executor: exec,
		config:   config,
	}
}

// String return name of workload.
func (y ycsb) String() string {
	return name
}

func (y ycsb) Populate() (err error) {
	populateCmd := y.buildPopulateCommand()

	taskHandle, err := y.executor.Execute(populateCmd)
	if err != nil {
		return err
	}

	_, err = taskHandle.Wait(0)
	if err != nil {
		return err
	}

	exitCode, err := taskHandle.ExitCode()
	if err != nil {
		return err
	}

	if exitCode != 0 {
		return errors.Errorf("Redis population exited with code: %d on command: %s", exitCode, populateCmd)
	}

	return nil
}

// TODO: Implementation
func (y ycsb) Tune(slo int) (qps int, achievedSLI int, err error) {
	return 0, 0, nil
}

func (y ycsb) Load(qps int, duration time.Duration) (executor.TaskHandle, error) {

	loadCommand := y.buildLoadCommand()

	taskHandle, err := y.executor.Execute(loadCommand)
	if err != nil {
		return nil, errors.Wrapf(err, "Execution of Ycsb Master Load failed. Command: %q", loadCommand)
	}

	return taskHandle, nil
}

func (y ycsb) buildPopulateCommand() string {

	cmd := fmt.Sprint(
		fmt.Sprintf("%s", y.config.PathToBinary),
		fmt.Sprint(" load redis -s"),
		fmt.Sprint(y.config.workloadCommand),
	)

	return cmd
}

func (y ycsb) buildLoadCommand() string {

	cmd := fmt.Sprint(
		fmt.Sprintf("%s", y.config.PathToBinary),
		fmt.Sprint(" run redis -s"),
		fmt.Sprint(y.config.workloadCommand),
	)

	return cmd
}
