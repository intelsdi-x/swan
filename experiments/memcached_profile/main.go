package main

import (
	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/swan/pkg/experiment"
)

type MemcachedSensitivityProfile struct {
	*experiment.Experiment
}

func NewMemcachedSensitivityProfile() *MemcachedSensitivityProfile {
	experiment := MemcachedSensitivityProfile{experiment.NewExperiment()}
	return &experiment
}

func (m *MemcachedSensitivityProfile) Init() {
	m.Name = "Memcached Sensitivity Profile"
	// Default experiment length in seconds
	m.Duration = 30
	// Expected SLO: 99%ile latency 5ms
	m.TargetSlo99pUs = 5000

	m.FindQPSAndLoadPoints()

	// Add phases.
	m.AddBaselinePhase(MemcachedBaselinePhase{})
	m.AddPhase(MemcachedWithL1InstructionPressurePhase{})
}

func (m *MemcachedSensitivityProfile) FindQPSAndLoadPoints() {
	log.Debug("Tuning phase. Finding QPS for ", m.TargetSlo99pUs, " SLO")
	// TODO(bplotka): Find QPSs from SLO

	// Let's assume for test 100:
	targetQPS := uint(100)
	m.InitLoadPoints(targetQPS)
}

// Baseline Phase
type MemcachedBaselinePhase struct {}

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

// L1InstructionPressure Test
type MemcachedWithL1InstructionPressurePhase struct {}

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
	experiment := NewMemcachedSensitivityProfile()
	experiment.Init()
	experiment.Run()

}
