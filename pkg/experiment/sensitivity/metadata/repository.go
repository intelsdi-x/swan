package metadata

// Repository is a interface for persisting Experiment to data store and fetching Experiment from data store.
type Repository interface {
	// Save persists experiment metadata to data store.
	Save(Experiment) error
	// Fetch retrieves experiment metadata from data store.
	Fetch(experiment string) (Experiment, error)
}
