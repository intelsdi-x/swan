package main

import (
	"fmt"
	"os"
	"strconv"

	swan "github.com/intelsdi-x/swan/pkg/experiment"
)

var phase_run_called int

type SamplePhase struct {
	name    string
	reps    int
	called  *int
	results []float64
}

func (p SamplePhase) Name() string {
	return p.name
}

func (p SamplePhase) Repetitions() int {
	return p.reps
}

func (p SamplePhase) Run() (float64, error) {
	(*p.called)++
	//Create log/output in current directory
	file, err := os.Create(p.name + "__" + strconv.FormatInt(int64(*p.called), 10) + ".log")
	if err != nil {
		return 0, err
	}
	message := "Sample output form phase " + p.name + "\n"
	file.WriteString(message)
	message = "Returning value: " +
		strconv.FormatFloat(p.results[(*p.called)-1], 'f', 4, 64) + "\n"
	file.WriteString(message)
	file.Close()
	return p.results[(*p.called)-1], nil
}

func main() {

	var phases []swan.Phase

	expConf := swan.ExperimentConfiguration{
		MaxVariance:      1,
		WorkingDirectory: "/tmp",
	}

	samplePhase := &SamplePhase{
		name:    "exp_ph_01",
		reps:    5,
		results: []float64{12.5, 1.9, 0, 4.7, 9.0},
		called:  &phase_run_called,
	}

	phases = append(phases, samplePhase)

	exp, err := swan.NewExperiment(expConf, phases)

	if err != nil {

	}
	fmt.Println("Starting new Experiment")
	exp.Run()
	fmt.Println("Done")
}
