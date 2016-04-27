package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/intelsdi-x/swan/pkg/executor"
)

var (
	tuneResults []float64
	targetQPS   int
)

func addTuneResult(result float64) {
	tuneResults = append(tuneResults, result)
}

func getTuneResults() []float64 {
	return tuneResults
}

////////////////////////////////////////////////////////////////////////////////
type TunePhase struct {
	name     string
	reps     int
	duration int
}

func (p TunePhase) Name() string {
	return p.name
}

func (p TunePhase) Repetitions() int {
	return p.reps
}

func (p TunePhase) Run() error {
	memcached := executor.NewLocal()
	mutilate := executor.NewLocal()

	fmt.Printf("Phase %s starting...\n", p.name)

	fmt.Println(" Starting memcached")
	memcachedTask, err := memcached.Execute("/tmp/MemcachedExperiment/memcached -u root -t 1")
	if err != nil {
		fmt.Println("->/Failed")
		fmt.Println(err)
		return err
	}
	taskState, taskStatus := memcachedTask.Status()
	if taskState != executor.RUNNING {
		fmt.Println("Task state (memcached) is not RUNNING!")
		return errors.New("Local task not running")
	}
	if taskStatus != nil {
		fmt.Println("Memcached task status is nil")
		return errors.New("Invalid command status")
	}

	//Wait 4 seconds
	memcachedTask.Wait(4000)

	fmt.Println("  Starting mutilate")
	mutilateStr := fmt.Sprintf("/tmp/MemcachedExperiment/mutilate --search 99:9 -s 127.0.0.1 -t %d -T 4 -c 64", p.duration)
	mutilateTask, err := mutilate.Execute(mutilateStr)
	if err != nil {
		fmt.Println(err)
		return err
	}
	taskState, taskStatus = mutilateTask.Status()
	if taskState != executor.RUNNING {
		fmt.Println("Task state (mutilate) is not RUNNING!")
		return errors.New("Local task not running")
	}
	if taskStatus != nil {
		fmt.Println("Mutilate task status is nil")
		return errors.New("Invalid command status")
	}

	// Task should run now. Wait 5 seconds - it won't terminate
	isTerminated := mutilateTask.Wait(0)
	if isTerminated != true {
		return errors.New("Task still running! Should not")
	}
	fmt.Printf("  Mutilate ended\n")
	err = out2file("mutilate-", mutilateTask)
	if err != nil {
		fmt.Println(err)
	}
	_, taskStatus = mutilateTask.Status()
	_, qps := mutilateGetResult(taskStatus.Stdout)
	addTuneResult(qps)

	err = memcachedTask.Stop()
	if err != nil {
		fmt.Println("Failed to stop memcached")
		fmt.Println(err)
		return err
	}

	taskState, taskStatus = memcachedTask.Status()
	if taskState != executor.TERMINATED {
		fmt.Println("memcached is not in terminated state")
		return errors.New("Task not terminated")
	}
	fmt.Println(" Memcached killed")
	err = out2file("memcached-", memcachedTask)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Phase %s ended\n", p.name)
	return nil
}

////////////////////////////////////////////////////////////////////////////////
type TuneCalcPhase struct {
}

func (p TuneCalcPhase) Name() string {
	return "Tune Calculation"
}

func (p TuneCalcPhase) Repetitions() int {
	return 1
}

func (p TuneCalcPhase) Run() error {
	var buffer bytes.Buffer

	f, err := os.Create("output.txt")

	results := getTuneResults()

	str := "Starting Tune Calculation Phase\n"
	buffer.WriteString(str)
	fmt.Println(str)
	//f.WriteString(str)
	variance := variance(results)
	average := average(results)
	str = fmt.Sprintf("Average Value: %2.2f, variance : %2.2f\n", average, variance)
	buffer.WriteString(str)
	str = fmt.Sprintf("Values: %v\n", results)
	buffer.WriteString(str)
	buffer.WriteString("Tune Calculation Phase done")
	targetQPS = int(average)

	if err == nil {
		f.WriteString(buffer.String())
		f.Close()
	}
	fmt.Println(buffer.String())

	return nil
}
