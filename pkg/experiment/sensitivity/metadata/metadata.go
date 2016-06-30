package metadata

import "time"

// BaseExperiment common part of all structs that represent sensitivity profile experiment metadata.
type BaseExperiment struct {
	ID                string
	LoadDuration      time.Duration
	TuningDuration    time.Duration
	LCName            string
	LGName            string
	RepetitionsNumber int
	LoadPointsNumber  int
	SLO               int
}

// Experiment represents metadata of single sensitivity profile experiment experiment and is agnostic to data store.
type Experiment struct {
	BaseExperiment
	Phases []Phase
}

// BasePhase is common part of all structs that represent phase of sensitivity profile experiment metadata.
type BasePhase struct {
	ID           string
	LCParameters string
	LCIsolation  string
}

// Phase represents metadata of single Phase in sensitivity profile experiment and is agnostic to data store.
type Phase struct {
	BasePhase
	Aggressors   []Aggressor
	Measurements Measurements
}

// Aggressor represents metadata of phase aggressor in sensitivity profile experiment and is agnostic to data store.
type Aggressor struct {
	Name       string
	Parameters string
	Isolation  string
}

// Measurement represents metadata of single measurement in sensitivity profile experiment and is agnostic to data store.
type Measurement struct {
	Load         *int
	LoadPointQPS int
	LGParameters string
}

// Measurements represents slice of Measurement structs.
type Measurements []Measurement

// AddPhase adds a Phase to the Experiment.
func (e *Experiment) AddPhase(phase Phase) {
	e.Phases = append(e.Phases, phase)
}

// AddMeasurement adds a Measurement to the Phase.
func (p *Phase) AddMeasurement(measurement Measurement) {
	p.Measurements = append(p.Measurements, measurement)
}

// AddAggressor adds Aggressor to the Phase.
func (p *Phase) AddAggressor(aggressor Aggressor) {
	p.Aggressors = append(p.Aggressors, aggressor)
}

// Len implements sort.Interface.
func (m Measurements) Len() int {
	return len(m)
}

// Less implements sort.Interface.
func (m Measurements) Less(i, j int) bool {
	return *m[i].Load < *m[j].Load
}

// Swap implements sort.Interface.
func (m Measurements) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
