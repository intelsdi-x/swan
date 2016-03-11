package main

import (
	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/swan/pkg/experiment"
)

func FindQPSAndLoadPoints(targetSLO uint) uint {
	log.Debug("Tuning phase. Finding QPS for ", targetSLO, " SLO")
	// TODO(bplotka): Find QPSs from SLO
	// - create Cgroup topology
	// - run memcached
	// - run mutilate in find qps mode
	// - clean cgroups

	// TODO(bplotka): Move that to tests
	// Let's assume for testing purposes a 100 QPS our targetSLO:
	targetQPS := uint(100)

	return targetQPS
}

// Baseline Phase
type MemcachedBaselinePhase struct{}

func (m MemcachedBaselinePhase) GetBestEffortWorkloadName() string {
	return "None"
}

func (m MemcachedBaselinePhase) Run(stresserLoad uint) float64 {
	// TODO(bplotka): Baseline memcached:
	// - create Cgroup topology
	// - run memcached
	// - run mutilate
	// - clean cgroups
	// - return SLIs (Measurements)

	// TODO(bplotka): Move that to tests
	// Fake SLI measurement.
	return float64(stresserLoad * 10)
}

// L1InstructionPressure Test.
type MemcachedWithL1InstructionPressurePhase struct{}

func (m MemcachedWithL1InstructionPressurePhase) GetBestEffortWorkloadName() string {
	return "L1 Instruction Pressure"
}

func (m MemcachedWithL1InstructionPressurePhase) Run(stresserLoad uint) float64 {
	// TODO(bplotka): Run Measurement with memcached & antagonist:
	// - create Cgroup topology
	// - run memcached
	// - run aggressor (here: L1InstructionPressure)
	// - run mutilate
	// - clean cgroups
	// - return SLIs (Measurements)

	// TODO(bplotka): Move that to tests
	// Fake SLI measurement (+ 10 to all latencies ((: )
	return float64((stresserLoad * 10) + 10)
}

func main() {
	memcachedExperiment := experiment.NewExperiment()
	memcachedExperiment.Name = "Memcached Sensitivity Profile"
	// Default experiment length in seconds
	memcachedExperiment.Duration = 30
	// Expected SLO: 99%ile latency 5ms
	memcachedExperiment.TargetSlo99pUs = 5000

	memcachedExperiment.InitLoadPoints(
		FindQPSAndLoadPoints(memcachedExperiment.TargetSlo99pUs))

	// Add phases.
	memcachedExperiment.AddBaselinePhase(MemcachedBaselinePhase{})
	memcachedExperiment.AddPhase(MemcachedWithL1InstructionPressurePhase{})

	// Run experiment.
	memcachedExperiment.Run()
}
