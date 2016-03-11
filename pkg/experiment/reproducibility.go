package experiment

type ReproducibilityError struct{}

func (r*ReproducibilityError) Error() string {
	return "Variance of the run is too high"
}

type ReproducibilityValidator struct {
	Iterations uint
	VarianceAcceptanceThreshold float64

	ValidatedResults uint
	results []float64
}

func NewReproducibilityValidator(
		iterations uint, varianceAcceptanceThreshold float64) *ReproducibilityValidator {
	return &ReproducibilityValidator{iterations, varianceAcceptanceThreshold, 0, []float64{}}
}

func (r ReproducibilityValidator) NeedToGather() bool {
	return r.ValidatedResults < r.Iterations;
}

func (r *ReproducibilityValidator) GatherResult(result float64) {
	r.results = append(r.results, result)
	r.ValidatedResults++
}

func (r *ReproducibilityValidator) Validate() (float64, error) {
	// TODO(bplotka): Calculate variance from results.
	// Mocked value.
	variance := 0.2
	if variance > r.VarianceAcceptanceThreshold {
		return 0, &ReproducibilityError{}
	}

	return r.GetMeanResult(), nil
}

func (r *ReproducibilityValidator) GetMeanResult() float64 {
	mean := float64(0)
	for _, result := range r.results {
		mean += result
	}

	return mean / float64(len(r.results))
}