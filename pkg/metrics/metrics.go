package metrics

import (
	"strings"
	"time"
)

// Tags contains values which are identifying SwanMetrics.
// NOTE (squall0): For further encoding(e.g.: with JSON Marshaler), fields here must be exported
type Tags struct {
	ExperimentID string
	PhaseID      string
	RepetitionID int
}

// Compare is compering current instance of Tags structure with another one.
func (t *Tags) Compare(tag Tags) bool {
	if strings.Compare(t.ExperimentID, tag.ExperimentID) != 0 {
		return false
	}
	if strings.Compare(t.PhaseID, tag.PhaseID) != 0 {
		return false
	}
	if t.RepetitionID != tag.RepetitionID {
		return false
	}

	return true
}

// Metadata swan runtime metrics.
type Metadata struct {
	AggressorName       string `json:",omitempty"`
	AggressorParameters string `json:",omitempty"`

	LCIsolation  string `json:",omitempty"`
	LCName       string `json:",omitempty"`
	LCParameters string `json:",omitempty"`

	LGIsolation  string `json:",omitempty"`
	LGName       string `json:",omitempty"`
	LGParameters string `json:",omitempty"`

	LoadDuration   time.Duration `json:",omitempty,string"`
	TuningDuration time.Duration `json:",omitempty,string"`

	LoadPointsNumber  int `json:",omitempty,string"`
	RepetitionsNumber int `json:",omitempty,string"`

	QPS int `json:",omitempty,string"`
	SLO int `json:",omitempty,string"`

	TargetLoad int `json:",omitempty,string"`
}

// Swan contains Metadata of experiment and ID values.
type Swan struct {
	Tags    Tags
	Metrics Metadata
}

// New is a constructor for SwanMetrics structure.
func New(tags Tags, metrics Metadata) *Swan {
	return &Swan{
		Tags:    tags,
		Metrics: metrics,
	}
}
