package ycsb

import (
	"fmt"
	"time"
)

// CalculateWorkloadCommandParameters parses parameters from config and creates command.
func CalculateWorkloadCommandParameters(qps int, duration time.Duration, config *Config) {

	config.WorkloadOperationCount = int64(qps) * int64(duration.Seconds())

	config.workloadCommand = fmt.Sprint(
		fmt.Sprintf(" -p redis.host=%s", config.RedisHost),
		fmt.Sprintf(" -p redis.port=%d", config.RedisPort),
		fmt.Sprintf(" -p recordcount=%d", config.WorkloadRecordCount),
		fmt.Sprintf(" -p operationcount=%d", config.WorkloadOperationCount),
		fmt.Sprintf(" -p workload=%s", config.Workload),
		fmt.Sprintf(" -p readallfields=%t", config.WorkloadReadAllFields),
		fmt.Sprintf(" -p readproportion=%g", config.WorkloadReadProportion),
		fmt.Sprintf(" -p updateproportion=%g", config.WorkloadUpdateProportion),
		fmt.Sprintf(" -p scanproportion=%g", config.WorkloadScanProportion),
		fmt.Sprintf(" -p insertproportion=%g", config.WorkloadInsertProportion),
		fmt.Sprintf(" -p requestdistribution=%s", config.WorkloadRequestDistribution),
		fmt.Sprintf(" -target %d", qps),
	)

}
