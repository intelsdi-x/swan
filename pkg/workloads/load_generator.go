package workloads

// LoadGenerator launches stresser which generates load on specified workload.
type LoadGenerator interface {
	// Tune does the tuning phase which is a process of searching for a targetQPS
	// for given SLO.
	Tune(slo int, timeoutMs int) (targetQPS int, err error)

	// Load generates load on the specific workload with the defined loadPoint (number of QPS).
	Load(qps int, durationMs int) (sli int, err error)
}
