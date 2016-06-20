package sensitivity

import (
	"github.com/intelsdi-x/swan/pkg/experiment"
	"time"
)

// Metadata swan runtime metrics.
type Metadata struct {
	experiment.Metadata

	AggressorName       []string `json:",omitempty"`
	AggressorParameters []string `json:",omitempty"`
	AggressorIsolations []string `json:",omitempty"`

	LCIsolation  string `json:",omitempty"`
	LCName       string `json:",omitempty"`
	LCParameters string `json:",omitempty"`

	LGIsolation  string   `json:",omitempty"`
	LGName       []string `json:",omitempty"`
	LGParameters []string `json:",omitempty"`

	LoadDuration   time.Duration `json:",omitempty,string"`
	TuningDuration time.Duration `json:",omitempty,string"`

	LoadPointsNumber  int `json:",omitempty,string"`
	RepetitionsNumber int `json:",omitempty,string"`

	Load float64 `json:",omitempty,string"`
	QPS  float64 `json:",omitempty,string"`
}
