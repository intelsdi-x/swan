package metadata

// Repository is a interface for persisting Experiment to data store and fetching Experiment from data store
type Repository interface {
	Save(Experiment) error
	Fetch(experiment string) (Experiment, error)
}
