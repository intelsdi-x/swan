package sensitivity

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
	"github.com/intelsdi-x/swan/pkg/workloads"
	"github.com/montanaflynn/stats"
)

type tuningPhase struct {
	// Latency Sensitivity (Production) workload.
	pr workloads.Launcher
	// Workload Generator for Latency Sensitivity task.
	lgForPr workloads.LoadGenerator
	// Given Service Level Objective.
	SLO int
	// Number of repetitions
	repetitions uint

	// Results across repetitions.
	// Load which was achieved during experiment e.g QPS, RPS.
	loadResults []float64
	// Service Level Indicator which was achieved during experiment e.g latency in us
	sliResults []float64

	// Shared reference for TargetLoad needed for Measurement phases.
	TargetLoad *int
}

// Returns Phase name.
func (p *tuningPhase) Name() string {
	return "Tuning_Phase"
}

// Returns number of repetitions.
func (p *tuningPhase) Repetitions() uint {
	return p.repetitions
}

// Run runs a tuning phase to find the targetLoad.
func (p *tuningPhase) Run(phase.Session) error {
	prTask, err := p.pr.Launch()
	if err != nil {
		return err
	}
	defer prTask.Stop()
	defer prTask.Clean()

	achievedLoad, achievedSLI, err := p.lgForPr.Tune(p.SLO)
	if err != nil {
		return err
	}

	// Save results.
	p.sliResults = append(p.sliResults, float64(achievedSLI))
	p.loadResults = append(p.loadResults, float64(achievedLoad))

	return err
}

// Finalize is executed after all repetitions of given measurement.
func (p *tuningPhase) Finalize() error {
	// TODO: Check if the variance is not too high.
	// For need results from the tuning phase in further experiments, so we
	// don't use snap here.

	// Calculate average.
	targetLoad, err := stats.Mean(p.loadResults)
	*p.TargetLoad = int(targetLoad)
	if err != nil {
		p.TargetLoad = nil
		return err
	}
	logrus.Debug("Calculated targetLoad (QPS/RPS): ", targetLoad, " for SLO: ", p.SLO)

	return nil
}
