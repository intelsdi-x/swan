package experiment

//Fraction of target QPS
type LoadPoint struct {
	no    int // Index
	value int // Value
}

func (l LoadPoint) String() string {
	return "LoadPoint object not defined"
}

// Experiement Conditions:
// Parameters defined for each experiment. Include:
// * SLO
// * Time of experiment
// * Time of measurement
// * Number of measurements
// * Acceptable variance between measurements results
// * Isolation configuration

type Configuration struct {
	//Service Level Objective
	SLO int
	// Time of experiement [s]
	expr_time int
	// Time of single measurement [s]
	meas_time int
	// How many time repeat the measurement
	num_meas int
	// Acceptable variance between measurements [%]
	meas_variance int
	// Number of Load Points
	NumLoadPoints int
}

func (c Configuration) String() string {
	return "Configuration object not defined"
}
