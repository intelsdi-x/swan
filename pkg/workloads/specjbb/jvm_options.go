package specjbb

import (
	"fmt"

	"github.com/intelsdi-x/swan/pkg/conf"
)

var (
	// JVMHeapMemoryGBs specifies amount of heap memory available to JVM.
	JVMHeapMemoryGBs = conf.NewIntFlag("specjbb_jvm_heap_size", "Size of JVM heap memory in gigabytes", 10)
)

// JVMOptions is group of options used to configure JVM for SPECjbb.
type JVMOptions struct {
	JVMHeapMemoryGBs  int
	ParallelGCThreads int
}

// DefaultJVMOptions returns sane JVMOptions.
func DefaultJVMOptions() JVMOptions {
	return JVMOptions{
		JVMHeapMemoryGBs:  JVMHeapMemoryGBs.Value(),
		ParallelGCThreads: 8,
	}
}

// GetJVMOptions returns string with JVM Options based on JVMOptions structure.
func (j JVMOptions) GetJVMOptions() string {
	return fmt.Sprint(
		" -server", // Compilation takes more time but offers additional optimizations
		fmt.Sprintf(" -Xms%dg -Xmx%dg", j.JVMHeapMemoryGBs, j.JVMHeapMemoryGBs), // Allocate whole heap available; docs: For best performance, set -Xms to the same size as the maximum heap size
		" -XX:NativeMemoryTracking=summary",                                     // Memory monitoring purposes
		" -XX:+UseParallelOldGC ",                                               // Parallel garbage collector for Old & Young generation.
		fmt.Sprintf(" -XX:ParallelGCThreads=%d", j.ParallelGCThreads),           // Sets the value of n to the number of logical processors. The value of n is the same as the number of logical processors up to a value of 8.
		fmt.Sprintf(" -XX:ConcGCThreads=%d", j.ParallelGCThreads/2),             // Currently half of PGCThreads.
		" -XX:InitiatingHeapOccupancyPercent=80",                                // Using more memory then default 45% before GC kicks in
		" -XX:MaxGCPauseMillis=100",                                             // Maximum garbage collection pause.
		" -XX:+AlwaysPreTouch ",                                                 // Touch & zero whole heap memory on initialization.
	)
}
