package phase

const (
	// ExperimentKey defines the key for Snap tag.
	ExperimentKey = "swan_experiment"
	// PhaseKey defines the key for Snap tag.
	PhaseKey = "swan_phase"
	// RepetitionKey defines the key for Snap tag.
	RepetitionKey = "swan_repetition"

	// LoadPointQPSKey defines the key for Snap tag.
	LoadPointQPSKey = "swan_loadpoint_qps"
	// AggressorNameKey defines the key for Snap tag.
	AggressorNameKey = "swan_aggressor_name"
)

// Session consists of data which make each phase unique.
type Session struct {
	ExperimentID  string
	PhaseID       string
	RepetitionID  int
	LoadPointQPS  int
	AggressorName string
}
