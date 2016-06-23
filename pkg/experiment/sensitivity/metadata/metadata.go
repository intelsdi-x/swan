package metadata

import (
	"time"

	"github.com/intelsdi-x/swan/pkg/experiment"
)

// Experiment represents metadata of single sensitivity experiment
type Experiment struct {
	experiment.Metadata

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

// Phase represents metadata of single Phase in sensitivity experiment
type Phase struct {
	ID                  string
	LCParameters        string
	LCIsolation         string
	AggressorNames      []string
	AggressorParameters []string
	AggressorIsolations []string
	Measurements        []Measurement
}

// Measurement represents metadata of single measurement in sensitivity experiment
type Measurement struct {
	Load         float64
	LoadPointQPS float64
	LGParameters []string
}

// AddPhase adds a Phase to the Experiment
func (e *Experiment) AddPhase(phase Phase) {
	e.Phases = append(e.Phases, phase)
}

// AddMeasurement adds a Measurement to the Experiment
func (p *Phase) AddMeasurement(measurement Measurement) {
	p.Measurements = append(p.Measurements, measurement)
}

// Metadata swan runtime metrics.
/*type Metadata struct {
	experiment.Metadata

	LCName string `json:",omitempty"`

	LGIsolation string   `json:",omitempty"`
	LGName      []string `json:",omitempty"`

	LoadDuration   time.Duration `json:",omitempty,string"`
	TuningDuration time.Duration `json:",omitempty,string"`

	LoadPointsNumber  int `json:",omitempty,string"`
	RepetitionsNumber int `json:",omitempty,string"`

	Phases []Phase
}

type Phase struct {
	LCIsolation  string   `json:",omitempty"`
	LCParameters string   `json:",omitempty"`
	LGParameters []string `json:",omitempty"`

	AggressorName       []string `json:",omitempty"`
	AggressorParameters []string `json:",omitempty"`
	AggressorIsolations []string `json:",omitempty"`

	Load float64 `json:",omitempty,string"`
	QPS  float64 `json:",omitempty,string"`
}*/
