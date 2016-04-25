package main

import (
	"errors"
	"fmt"
	"os"

	executor "github.com/intelsdi-x/swan/pkg/executor"
	swan "github.com/intelsdi-x/swan/pkg/experiment"
)

type localPhase struct {
	name string
	reps int
}

func (p localPhase) Name() string {
	return p.name
}

func (p localPhase) Repetitions() int {
	return p.reps
}

func (p localPhase) Run() error {
	l := executor.NewLocal()

	task, err := l.Execute("/usr/bin/vmstat 1")
	if err != nil {
		return err
	}
	taskState, taskStatus := task.Status()
	if taskState != executor.RUNNING {
		return errors.New("Local task not running")
	}
	if taskStatus != nil {
		return errors.New("Invalid command status")
	}
	// Task should run now. Wait 5 seconds - it won't terminate
	isTerminated := task.Wait(5000)
	if isTerminated != false {
		return errors.New("Task terminated! Should not")
	}
	err = task.Stop()
	if err != nil {
		return err
	}

	taskState, taskStatus = task.Status()
	if taskState != executor.TERMINATED {
		return errors.New("Task not terminated")
	}

	f, err := os.Create("./stdout.txt")
	if err != nil {
		return err
	}

	n, err := f.WriteString(taskStatus.Stdout)
	if err != nil {
		return err
	}
	if n != len(taskStatus.Stdout) {
		return errors.New("Write error")
	}
	f.Close()

	f, err = os.Create("./stderr.txt")
	if err != nil {
		return err
	}

	n, err = f.WriteString(taskStatus.Stderr)
	if err != nil {
		return err
	}
	if n != len(taskStatus.Stderr) {
		return errors.New("Write error")
	}
	f.Close()

	return err
}

func localExperiment() {

	var phases []swan.Phase

	localPhase := &localPhase{
		name: "vmstat phase",
		reps: 3,
	}

	phases = append(phases, localPhase)

	exp, err := swan.NewExperiment("SampleExperiment", phases, "/tmp")

	if err != nil {

	}
	fmt.Println("Starting Local Experiment")
	exp.Run()
	fmt.Println("Done")
}
