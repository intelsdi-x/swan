package main

import (
	"errors"
	"fmt"

	"github.com/intelsdi-x/swan/pkg/executor"
)

//Format: [numberOfLoadPoints][Repetition]
var baselineResults [][]float64

func addBaselineResult(loadPointIndex int, result float64) {
	baselineResults[loadPointIndex] =
		append(baselineResults[loadPointIndex], result)
}

////////////////////////////////////////////////////////////////////////////////
type BaselinePhase struct {
	name       string
	reps       int
	duration   int
	loadPoints int
}

func (p BaselinePhase) Name() string {
	return p.name
}

func (p BaselinePhase) Repetitions() int {
	return p.reps
}

func (p BaselinePhase) Run() error {
	if len(baselineResults) == 0 {
		baselineResults = make([][]float64, p.loadPoints)
	}

	fmt.Printf("Phase %s starting...\n", p.name)
	loadPoint := int(targetQPS / (p.loadPoints + 1))
	for i := 1; i <= p.loadPoints; i++ {
		fmt.Printf(" Starting loadpoint %d/%d\n", i*loadPoint, targetQPS)
		latency, err := runLoadPoint(i*loadPoint, p.duration)
		if err != nil {
			return err
		}
		addBaselineResult(i-1, latency)
	}
	fmt.Printf("Phase %s Ended\n", p.name)
	return nil
}

func runLoadPoint(loadPoint int, duration int) (float64, error) {
	memcached := executor.NewLocal()
	mutilate := executor.NewLocal()
	fmt.Printf(" Given loadpoint %d qps\n", loadPoint)
	fmt.Println(" Starting memcached")
	memcachedTask, err := memcached.Execute("/tmp/MemcachedExperiment/memcached -u root")
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	taskState, taskStatus := memcachedTask.Status()
	if taskState != executor.RUNNING {
		return 0, errors.New("Local task not running")
	}
	if taskStatus != nil {
		return 0, errors.New("Invalid command status")
	}

	//Wait 4 seconds
	memcachedTask.Wait(4000)

	fmt.Println("  Starting mutilate")
	mutilateStr := fmt.Sprintf("/tmp/MemcachedExperiment/mutilate -s 127.0.0.1 -t %d -q %d -T 4 -c 64", duration, loadPoint)
	mutilateTask, err := mutilate.Execute(mutilateStr)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	taskState, taskStatus = mutilateTask.Status()
	if taskState != executor.RUNNING {
		return 0, errors.New("Local task not running")
	}
	if taskStatus != nil {
		return 0, errors.New("Invalid command status")
	}

	isTerminated := mutilateTask.Wait(0)
	if isTerminated != true {
		return 0, errors.New("Task still running! Should not")
	}
	prefix := fmt.Sprintf("mutilate-%d-", loadPoint)
	err = out2file(prefix, mutilateTask)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("  Mutilate ended")
	fmt.Println("  Parsing mutilate output")
	_, taskStatus = mutilateTask.Status()
	latency, _ := mutilateGetResult(taskStatus.Stdout)

	err = memcachedTask.Stop()
	if err != nil {
		fmt.Println("Failed to stop memcached")
		fmt.Println(err)
		return 0, err
	}

	taskState, taskStatus = memcachedTask.Status()
	if taskState != executor.TERMINATED {
		fmt.Println("memcached is not in terminated state")
		return 0, errors.New("Task not terminated")
	}
	prefix = fmt.Sprintf("memcached-%d-", loadPoint)
	err = out2file(prefix, memcachedTask)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(" Memcached killed")
	fmt.Printf(" Resulting latency: %.2f\n", latency)
	return latency, err
}

////////////////////////////////////////////////////////////////////////////////
type BaselinePrintPhase struct {
}

func (p BaselinePrintPhase) Name() string {
	return "Baseline Print Result"
}

func (p BaselinePrintPhase) Repetitions() int {
	return 1
}

func (p BaselinePrintPhase) Run() error {

	fmt.Printf("#loadpoint: result1 result2 result3 ...\n")
	loadPoint := int(targetQPS / len(baselineResults))
	for i := 0; i < len(baselineResults); i++ {
		fmt.Printf("%d| %d ", i, loadPoint*(1+i))
		for j := 0; j < len(baselineResults[i]); j++ {
			fmt.Printf("| %.2f ", baselineResults[i][j])
		}
		average := average(baselineResults[i])
		variance := variance(baselineResults[i])
		fmt.Printf(" |average=%.2f |variance=%.2f|\n", average, variance)
	}
	return nil
}
