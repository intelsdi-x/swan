package main

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	executor "github.com/intelsdi-x/swan/pkg/executor"
	swan "github.com/intelsdi-x/swan/pkg/experiment"
)

var (
	tuneResults     []float64
	baselineResults []float64
	targetQPS       float64
)

func addTuneResult(result float64) {
	tuneResults = append(tuneResults, result)
}

func getTuneResults() []float64 {
	return tuneResults
}

func addBaselineResult(result float64) {
	baselineResults = append(baselineResults, result)

}

func getBaselineResults() []float64 {
	return baselineResults
}

////////////////////////////////////////////////////////////////////////////////
type TunePhase struct {
	name string
	reps int
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
	mutilateTask, err := mutilate.Execute("/tmp/MemcachedExperiment/mutilate --search 99:9 -s 127.0.0.1 -t 1 -T 4 -c 64")
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
	addTuneResult(mutilateGetResult(taskStatus.Stdout))

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
	name      string
	reps      int
	tunephase string
	tuneReps  int
}

func (p TuneCalcPhase) Name() string {
	return p.name
}

func (p TuneCalcPhase) Repetitions() int {
	return p.reps
}

func (p TuneCalcPhase) Run() error {

	results := getTuneResults()

	fmt.Println("Starting Tune Calculation Phase")
	variance := variance(results)
	average := average(results)
	fmt.Printf("Average Value: %2.2f, variance : %2.2f\n", average, variance)
	fmt.Printf("Values: %v\n", results)

	targetQPS = average
	return nil
}

////////////////////////////////////////////////////////////////////////////////
type BaselinePhase struct {
	name string
	reps int
}

func (p BaselinePhase) Name() string {
	return p.name
}

func (p BaselinePhase) Repetitions() int {
	return p.reps
}

func (p BaselinePhase) Run() error {
	memcached := executor.NewLocal()
	mutilate := executor.NewLocal()

	fmt.Printf("Phase %s starting...\n", p.name)
	fmt.Println(" Starting memcached")
	memcachedTask, err := memcached.Execute("/tmp/MemcachedExperiment/memcached -u root -t 1")
	if err != nil {
		fmt.Println(err)
		return err
	}
	taskState, taskStatus := memcachedTask.Status()
	if taskState != executor.RUNNING {
		return errors.New("Local task not running")
	}
	if taskStatus != nil {
		return errors.New("Invalid command status")
	}

	//Wait 4 seconds
	memcachedTask.Wait(4000)

	fmt.Println("  Starting mutilate")
	mutilateTask, err := mutilate.Execute("/tmp/MemcachedExperiment/mutilate --search 99:9 -s 127.0.0.1 -t 1 -T 4 -c 64")
	if err != nil {
		fmt.Println(err)
		return err
	}
	taskState, taskStatus = mutilateTask.Status()
	if taskState != executor.RUNNING {
		return errors.New("Local task not running")
	}
	if taskStatus != nil {
		return errors.New("Invalid command status")
	}

	// Task should run now. Wait 5 seconds - it won't terminate
	isTerminated := mutilateTask.Wait(0)
	if isTerminated != true {
		return errors.New("Task still running! Should not")
	}
	err = out2file("mutilate-", mutilateTask)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("  Mutilate ended")
	fmt.Println("  Parsing mutilate output")
	_, taskStatus = mutilateTask.Status()
	addTuneResult(mutilateGetResult(taskStatus.Stdout))

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
	err = out2file("memcached-", memcachedTask)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(" Memcached killed")
	fmt.Printf("Phase %s Ended\n", p.name)
	return err
}

func out2file(prefix string, task executor.Task) error {
	_, taskStatus := task.Status()

	f, err := os.Create("./" + prefix + "stdout.txt")
	if err != nil {
		return err
	}

	n, err := f.WriteString(taskStatus.Stdout)
	if err != nil {
		f.Close()
		return err
	}
	if n != len(taskStatus.Stdout) {
		f.Close()
		return errors.New("Write error")
	}
	f.Close()

	f, err = os.Create("./" + prefix + "stderr.txt")
	if err != nil {
		return err
	}

	n, err = f.WriteString(taskStatus.Stderr)
	if err != nil {
		f.Close()
		return err
	}
	if n != len(taskStatus.Stderr) {
		f.Close()
		return errors.New("Write error")
	}
	f.Close()

	return err
}

func mutilateGetResult(output string) (latency float64) {
	var err error
	scanner := bufio.NewScanner(strings.NewReader(output))

	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := strings.Split(scanner.Text(), " ")
		if len(line) == 7 {
			if strings.Contains(scanner.Text(), "Total QPS") {
				latency, err = strconv.ParseFloat(line[3], 64)
				if err != nil {
					return 0
				}
				return latency
			}
		}
	}
	return latency
}

func average(x []float64) float64 {
	var avr float64
	for _, val := range x {
		avr += val
	}
	avr /= float64(len(x))
	return avr
}
func variance(x []float64) float64 {

	avr := average(x)
	variance := float64(0)
	for _, val := range x {
		variance += math.Sqrt(math.Abs(avr - val))
	}
	variance /= float64(len(x))
	return variance
}

func main() {
	var phases []swan.Phase

	tunePhase := &TunePhase{
		name: "Tuning",
		reps: 3,
	}

	tuneCalcPhase := &TuneCalcPhase{
		name: "Tune Calculation Phase",
		reps: 1,
	}

	baselinePhase := &BaselinePhase{
		name: "Baseline",
		reps: 3,
	}

	phases = append(phases, tunePhase)
	phases = append(phases, tuneCalcPhase)
	phases = append(phases, baselinePhase)

	exp, err := swan.NewExperiment("MemcachedExperiment", phases, "/tmp")
	if err != nil {
		return
	}
	fmt.Println("Starting Memcached Experiment")
	exp.Run()
	fmt.Println("Done")
}
