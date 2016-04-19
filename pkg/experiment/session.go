package experiment

// Session defines an execution of an experiment: a unique identifier, where to
// find the experiment files, etc.
type Session struct {
	Name    string
	WorkDir string
}
