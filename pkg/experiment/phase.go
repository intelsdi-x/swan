package experiment

type Phase interface {
	Name() string
	Run() (float64, error)
	Repetitions() int
}
