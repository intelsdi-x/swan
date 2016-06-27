package metadata

import "time"

// Experiment represents metadata of single sensitivity experiment.
type Experiment struct {
	ID                string
	LoadDuration      time.Duration
	TuningDuration    time.Duration
	LcName            string
	LgNames           []string
	RepetitionsNumber int
	LoadPointsNumber  int
	SLO               int
	Phases            []Phase
}

// Phase represents metadata of single Phase in sensitivity experiment.
type Phase struct {
	ID                  string
	LCParameters        string
	LCIsolation         string
	AggressorNames      []string
	AggressorParameters []string
	AggressorIsolations []string
	Measurements        Measurements
}

// Measurement represents metadata of single measurement in sensitivity experiment.
type Measurement struct {
	Load         float64
	LoadPointQPS float64
	LGParameters []string
}

// Measurements represents slice of Measurement structs.
type Measurements []Measurement

// AddPhase adds a Phase to the Experiment.
func (e *Experiment) AddPhase(phase Phase) {
	e.Phases = append(e.Phases, phase)
}

// AddMeasurement adds a Measurement to the Experiment.
func (p *Phase) AddMeasurement(measurement Measurement) {
	p.Measurements = append(p.Measurements, measurement)
}

// Len implements sort.Interface.
func (m Measurements) Len() int {
	return len(m)
}

// Less implements sort.Interface.
func (m Measurements) Less(i, j int) bool {
	return m[i].Load < m[j].Load
}

// Swap implements sort.Interface.
func (m Measurements) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
