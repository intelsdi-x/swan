package experiment

type Workload struct {
	//
}

func (w Workload) String() string {
	return "Workload not defined"
}

func (w Workload) Run(lp *LoadPoint) (SLI, int) {
	return SLI{}, 0
}
