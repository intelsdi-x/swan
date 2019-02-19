package ycsb

import (
	"github.com/intelsdi-x/swan/pkg/executor"
	"time"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/workloads/redis"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"github.com/cloudflare/cfssl/log"
)

const (
	name = "YCSB"
	defaultPathToBinary = "ycsb"
	defaultWorkload = "com.yahoo.ycsb.workloads.CoreWorkload"
	defaultWorkloadRecordCount = 100000
	defaultWorkloadReadAllFields = true
	defaultWorkloadReadProportion = "0.5"
	defaultWorkloadUpdateProportion = "0.5"
	defaultWorkloadScanProportion = "0.0"
	defaultWorkloadInsertProportion = "0.0"
	defaultWorkloadRequestDistribution = "zipfian"
)

var (
	PathFlag = conf.NewStringFlag("ycsb_path", "Path to YCSB binary file.", defaultPathToBinary)
	WorkloadFlag = conf.NewStringFlag("ycsb_workload", "Name of YCSB workload", defaultWorkload)
	WorkloadRecordCountFlag = conf.NewIntFlag("ycsb_workload_recordcount", "Workload record count.", defaultWorkloadRecordCount)
	WorkloadReadAllFieldsFlag = conf.NewBoolFlag("ycsb_workload_readallfields", "Workload read all fields.", defaultWorkloadReadAllFields)
	WorkloadReadProportionFlag = conf.NewStringFlag("ycsb_workload_readproportion", "Workload read proportion.", defaultWorkloadReadProportion)
	WorkloadUpdateProportionFlag = conf.NewStringFlag("ycsb_workload_updateproportion", "Workload update proportion.", defaultWorkloadUpdateProportion)
	WorkloadScanProportionFlag = conf.NewStringFlag("ycsb_workload_scanproportion", "Workload scan proportion.", defaultWorkloadScanProportion)
	WorkloadInsertProportionFlag = conf.NewStringFlag("ycsb_workload_insertproportion", "Workload insert proportion.", defaultWorkloadInsertProportion)
	WorkloadRequestDistributionFlag = conf.NewStringFlag("ycsb_workload_requestdistribution", "Workload request distribution.", defaultWorkloadRequestDistribution)
)

type ycsb struct {
	executor executor.Executor
	config Config
}

type Config struct {
	PathToBinary 				string
	RedisHost					string
	RedisPort					int
	RedisClusterMode			bool // Redis cluster mode.
	Workload					string
	WorkloadRecordCount			int
	WorkloadOperationCount		int64
	WorkloadReadAllFields		bool
	WorkloadReadProportion		float64
	WorkloadUpdateProportion	float64
	WorkloadScanProportion		float64
	WorkloadInsertProportion	float64
	WorkloadRequestDistribution	string
	workloadCommand				string
}

func DefaultYcsbConfig() Config {

	workloadReadProportion, err := strconv.ParseFloat(WorkloadReadProportionFlag.Value(), 64)
	if err != nil {
		workloadReadProportion = 0.0
	}

	workloadUpdateProportion, err := strconv.ParseFloat(WorkloadUpdateProportionFlag.Value(), 64)
	if err != nil {
		workloadUpdateProportion = 0.0
	}

	workloadScanProportion, err := strconv.ParseFloat(WorkloadScanProportionFlag.Value(), 64)
	if err != nil {
		workloadScanProportion = 0.0
	}

	workloadInsertProportion, err := strconv.ParseFloat(WorkloadInsertProportionFlag.Value(), 64)
	if err != nil {
		workloadInsertProportion = 0.0
	}

	return Config{
		PathToBinary: 					PathFlag.Value(),
		RedisHost:						redis.IPFlag.Value(),
		RedisPort:						redis.PortFlag.Value(),
		RedisClusterMode:				redis.ClusterFlag.Value(),
		Workload:						WorkloadFlag.Value(),
		WorkloadRecordCount: 			WorkloadRecordCountFlag.Value(),
		WorkloadReadAllFields: 			WorkloadReadAllFieldsFlag.Value(),
		WorkloadReadProportion:			workloadReadProportion,
		WorkloadUpdateProportion:		workloadUpdateProportion,
		WorkloadScanProportion:			workloadScanProportion,
		WorkloadInsertProportion:		workloadInsertProportion,
		WorkloadRequestDistribution:	WorkloadRequestDistributionFlag.Value(),
	}
}

func New(exec executor.Executor, config Config) executor.LoadGenerator {
	return ycsb{
		executor: exec,
		config: config,
	}
}


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

	log.Infof("Redis population exited with code: %d on command: %s", exitCode, populateCmd)

	return nil
}

// TODO: Implementation
func (y ycsb) Tune(slo int) (qps int, achievedSLI int, err error) {
	return 0,0,nil
}

func (y ycsb) Load(qps int, duration time.Duration) (executor.TaskHandle, error) {

	y.calculateWorkloadCommandParameters(qps, duration)

	err := y.Populate()
	if err != nil {
		return nil, err
	}

	loadCommand := y.buildLoadCommand()

	taskHandle, err := y.executor.Execute(loadCommand)
	if err != nil {
		return nil, errors.Wrapf(err, "Execution of Ycsb Master Load failed. Command: %q", loadCommand)
	}

	_, err = taskHandle.Wait(0)
	if err != nil {
		return nil, err
	}

	exitCode, err := taskHandle.ExitCode()
	if err != nil {
		return nil, err
	}

	if exitCode != 0 {
		return nil, errors.Errorf("YCSB load exited with code: %d on command: %s", exitCode, loadCommand)
	}

	return taskHandle, nil
}

func (y *ycsb) calculateWorkloadCommandParameters(qps int, duration time.Duration) {

	y.config.WorkloadOperationCount= int64(qps) * int64(duration.Seconds())

	y.config.workloadCommand = fmt.Sprint(
		fmt.Sprintf(" -p redis.host=%s", y.config.RedisHost),
		fmt.Sprintf(" -p redis.port=%d", y.config.RedisPort),
		fmt.Sprintf(" -p recordcount=%d", y.config.WorkloadRecordCount),
		fmt.Sprintf(" -p operationcount=%d", y.config.WorkloadOperationCount),
		fmt.Sprintf(" -p workload=%s", y.config.Workload),
		fmt.Sprintf(" -p readallfields=%t", y.config.WorkloadReadAllFields),
		fmt.Sprintf(" -p readproportion=%g", y.config.WorkloadReadProportion),
		fmt.Sprintf(" -p updateproportion=%g", y.config.WorkloadUpdateProportion),
		fmt.Sprintf(" -p scanproportion=%g", y.config.WorkloadScanProportion),
		fmt.Sprintf(" -p insertproportion=%g", y.config.WorkloadInsertProportion),
		fmt.Sprintf(" -p requestdistribution=%s", y.config.WorkloadRequestDistribution),
		fmt.Sprintf(" -target %d", qps),
	)

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
