package experiment

import (
	log "github.com/Sirupsen/logrus"
)

func initLogging() {
	log.SetLevel(log.DebugLevel)
}

func preExperimentVerification() {
	log.Debug("Verification of the extended number of File descriptors in OS...")
	// TOOD(bplotka)

	log.Debug("Verification of the CPU governor being set to `performance` for OS...")
	// TOOD(bplotka)

}

type Experiment struct {
	// Name of the experiment.
	Name string
	// Duration of the experiment.
	Duration uint
	// Expected SLO (99%ile latency in us).
	TargetSlo99pUs uint

	// TODO(bplotka): Make number of these load points more dynamic
	// 5% (0 index), 10%, ... 95% (18 index) load Points. Currently they are in form of the RPS/QPS
	// TODO(bplotka): Move that to map for clear view?
	loadPoints [19]uint
	targetLoad uint

	baselinePhase       Phase
	baselineMeasurement *Measurement

	phases       []Phase
	measurements map[string]*Measurement
}

func NewExperiment() *Experiment {
	initLogging()
	preExperimentVerification()

	e := Experiment{}
	e.targetLoad = 0
	e.measurements = make(map[string]*Measurement)

	return &e
}

func (e *Experiment) InitLoadPoints(targetLoad uint) {
	e.targetLoad = targetLoad
	e.loadPoints = [19]uint{}

	// TODO(bplotka): Make number of these load points more dynamic
	// Create 5%, 10% ... 100% load points.
	for i := 0; i < 19; i++ {
		e.loadPoints[i] = uint(float64(targetLoad) * float64(i+1) * 0.05)
	}

	// TODO(bplotka): Move that to tests
	// Debug log to ensure that load points are calculated correctly.
	for i := 0; i < 19; i++ {
		log.Debug(e.loadPoints[i])
	}
}

func (e *Experiment) AddBaselinePhase(baselinePhase Phase) {
	// Register Baseline.
	e.baselineMeasurement = &Measurement{}
	e.baselinePhase = baselinePhase
}

func (e *Experiment) AddPhase(phase Phase) {
	// Register Phase.
	// TODO(bplotka) Assure uniqueness in the GetBestEffortWorloadName.
	e.measurements[phase.GetBestEffortWorkloadName()] = &Measurement{}
	e.phases = append(e.phases, phase)
}

func (e *Experiment) runPhaseFully(phase Phase, resultBucket *Measurement) {
	// Save result of measurement from one load Point.
	for i, loadPoint := range e.loadPoints {
		log.Debug("Running phase. Load: ", (i+1)*5, "% = ", loadPoint, " RPS/QPS")
		resultBucket.SliLoadPointLatencies99pUs[i] =
			phase.Run(loadPoint)
	}
}

func (e *Experiment) saveMeasurements() {
	// Save result of measurement to files.
	// TODO(bplotka) Do saving to a file.
	// TODO(bplotka) Move that to test.
	log.Debug("Saving measurements from baseline = ", e.baselineMeasurement.SliLoadPointLatencies99pUs)

	for antagonist, meas := range e.measurements {
		log.Debug("Saving measurements from phase with ", antagonist,
			" = ", meas.SliLoadPointLatencies99pUs)
	}
}

func (e *Experiment) Run() {
	if e.targetLoad == 0 {
		log.Error("Load points are not initialised!")
		return
	}

	log.Info("Running experiment '", e.Name, "'")

	// Run baseline.
	log.Info("-----Running baseline phase-----")
	e.runPhaseFully(e.baselinePhase, e.baselineMeasurement)
	log.Info("-----End of phase baseline phase------")

	// Run rest of the phases.
	for _, phase := range e.phases {
		log.Info("-----Running phase with ", phase.GetBestEffortWorkloadName(), "------")
		e.runPhaseFully(phase, e.measurements[phase.GetBestEffortWorkloadName()])
		log.Info("-----End of phase with ", phase.GetBestEffortWorkloadName(), "------")
	}

	// Save results to the file.
	e.saveMeasurements()
}
