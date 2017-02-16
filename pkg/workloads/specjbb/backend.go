package specjbb

import (
	"fmt"

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/pkg/errors"
)

const (
	name         = "SPECjbb Backend"
	backendJvmID = "specjbbbackend1"
)

var (
	// PathToBinaryForHpFlag specifies path to a SPECjbb2015 jar file for hp job.
	PathToBinaryForHpFlag = conf.NewStringFlag("specjbb_path_hp",
		"Path to SPECjbb jar for high priority job (backend)",
		"/usr/share/specjbb/specjbb2015.jar")
	// PathToPropsFileForHpFlag specifies path to a SPECjbb2015 properties file for hp job.
	PathToPropsFileForHpFlag = conf.NewStringFlag("specjbb_props_path_hp",
		"Path to SPECjbb properties file for high priority job (backend)",
		"/usr/share/specjbb/config/specjbb2015.props")

	// JVMHeapMemoryGBs specifies amount of heap memory available to JVM.
	JVMHeapMemoryGBs = conf.NewIntFlag("specjbb_jvm_heap_size", "Size of JVM heap memory in gigabytes", 10)
)

// BackendConfig is a config for a SPECjbb2015 Backend,
type BackendConfig struct {
	PathToBinary      string
	ControllerAddress string // ControllerAddress is an address of a SPECjbb controller component ("-Dspecjbb.controller.host=")
	JvmID             string // JvmId is an ID of a JVM dedicated for a Backend (-J <jvmid>)
	JVMHeapMemoryGBs  int    // JVMHeapMemoryGBs is number of GBs available for JVM heap.
	Parallelism       int    // Amount of threads in ForkJoinPool.
}

// DefaultSPECjbbBackendConfig is a constructor for BackendConfig with default parameters.
func DefaultSPECjbbBackendConfig() BackendConfig {
	return BackendConfig{
		PathToBinary:      PathToBinaryForHpFlag.Value(),
		ControllerAddress: ControllerAddress.Value(),
		JvmID:             backendJvmID,
		JVMHeapMemoryGBs:  JVMHeapMemoryGBs.Value(),
		Parallelism:       8,
	}
}

// Backend is a launcher for the SPECjbb2015 Backend.
type Backend struct {
	exec executor.Executor
	conf BackendConfig
}

// NewBackend is a constructor for Backend.
func NewBackend(exec executor.Executor, config BackendConfig) Backend {
	return Backend{
		exec: exec,
		conf: config,
	}
}

func (b Backend) buildCommand() string {
	// See: https://intelsdi.atlassian.net/wiki/display/SCE/SpecJBB+experiment+tuning
	return fmt.Sprint("java -jar",
		" -server", // Compilation takes more time but offers additional optimizations

		fmt.Sprintf(" -Djava.util.concurrent.ForkJoinPool.common.parallelism=%d", b.conf.Parallelism), // Amount of threads equal to amount of hyper threads

		fmt.Sprintf(" -Xms%dg -Xmx%dg", b.conf.JVMHeapMemoryGBs, b.conf.JVMHeapMemoryGBs), // Allocate whole heap available; docs: For best performance, set -Xms to the same size as the maximum heap size
		" -XX:NativeMemoryTracking=summary",                                               // Memory monitoring purposes
		" -XX:+UseParallelGC",                                                             // Parallel garbage collector
		fmt.Sprintf(" -XX:ParallelGCThreads=%d", b.conf.Parallelism),                      // Sets the value of n to the number of logical processors. The value of n is the same as the number of logical processors up to a value of 8.
		fmt.Sprintf(" -XX:ConcGCThreads=%d", b.conf.Parallelism/2),                        // Currently half of PGCThreads.
		" -XX:InitiatingHeapOccupancyPercent=80",                                          // Using more memory then default 45% before GC kicks in
		" -XX:MaxGCPauseMillis=100",                                                       //Sets a target value for desired maximum pause time. The default value is 200 milliseconds. The specified value does not adapt to your heap size.

		ControllerHostProperty, b.conf.ControllerAddress,
		" ", b.conf.PathToBinary,
		" -m backend",
		" -G GRP1",
		" -J ", b.conf.JvmID,
		" -p ", PathToPropsFileForHpFlag.Value(),
	)
}

// Launch starts the Backend component. It returns a Task Handle instance.
// Error is returned when Launcher is unable to start a job.
func (b Backend) Launch() (executor.TaskHandle, error) {
	task, err := b.exec.Execute(b.buildCommand())
	if err != nil {
		return nil, errors.Wrapf(err, "launch of SPECjbb backend failed. command: %q", b.buildCommand())
	}
	return task, nil
}

// Name returns human readable name for job.
func (b Backend) Name() string {
	return name
}
