package specjbb

import (
	"fmt"
)

type JVMOptions struct {
	Parallelism      int
	JVMHeapMemoryGBs int
}

func DefaultJVMOptions() JVMOptions {
	return JVMOptions{
		Parallelism:      8,
		JVMHeapMemoryGBs: 8,
	}
}

func (j JVMOptions) GetJVMOptions() string {
	return fmt.Sprintf(
		" -server", // Compilation takes more time but offers additional optimizations
		fmt.Sprintf(" -Djava.util.concurrent.ForkJoinPool.common.parallelism=%d", j.Parallelism), // Amount of threads equal to amount of hyper threads
		fmt.Sprintf(" -Xms%dg -Xmx%dg", j.JVMHeapMemoryGBs, j.JVMHeapMemoryGBs),                  // Allocate whole heap available; docs: For best performance, set -Xms to the same size as the maximum heap size
		" -XX:NativeMemoryTracking=summary",                                                      // Memory monitoring purposes
		" -XX:+UseParallelGC",                                                                    // Parallel garbage collector
		fmt.Sprintf(" -XX:ParallelGCThreads=%d", j.Parallelism),                                  // Sets the value of n to the number of logical processors. The value of n is the same as the number of logical processors up to a value of 8.
		fmt.Sprintf(" -XX:ConcGCThreads=%d", j.Parallelism/2),                                    // Currently half of PGCThreads.
		" -XX:InitiatingHeapOccupancyPercent=80",                                                 // Using more memory then default 45% before GC kicks in
		" -XX:MaxGCPauseMillis=100",
	)
}

//Sets a target value for desired maximum pause time. The default value is 200 milliseconds. The specified value does not adapt to your heap size.
