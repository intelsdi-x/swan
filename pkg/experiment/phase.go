package experiment

type Phase interface {
	// In case of tests with antagonist or BE task workload name needs to be specified.
	GetBestEffortWorkloadName() string
	// Experiment struct will invoke this method for each loadPoint.
	Run(stresserLoad uint) float64
}
