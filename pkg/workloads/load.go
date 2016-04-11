package workloads

// LoadGenerator launches stresser which generates load on specified workload.
type LoadGenerator interface {
	// Tune does the tuning phase. With the given SLO.
	Tune(slo int, timeoutMs int) (targetQPS int, error)

	// Load generates load on specific workload.
	Load(qps int, durationMs int) (sli int, error)
}
