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

func mutilateGetResult(output string) (latency, qps float64) {
	var err error
	scanner := bufio.NewScanner(strings.NewReader(output))

	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := strings.Fields(scanner.Text())
		switch {
		case len(line) == 7:
			if strings.Contains(scanner.Text(), "Total QPS") {
				qps, err = strconv.ParseFloat(line[3], 64)
				if err != nil {
					return 0, 0
				}
			}
		case len(line) == 9:
			if line[0] == "read" {
				latency, err = strconv.ParseFloat(line[8], 64)
				if err != nil {
					return 0, 0
				}
			}
		}
	}
	return latency, qps
}

func main() {
	var phases []swan.Phase

	tunePhase := &TunePhase{
		name:     "Tuning",
		reps:     3,
		duration: 10,
	}

	tuneCalcPhase := &TuneCalcPhase{}

	baselinePhase := &BaselinePhase{
		name:       "Baseline",
		reps:       3,
		duration:   10,
		loadPoints: 5,
	}

	baselinePrintPhase := &BaselinePrintPhase{}

	phases = append(phases, tunePhase)
	phases = append(phases, tuneCalcPhase)
	phases = append(phases, baselinePhase)
	phases = append(phases, baselinePrintPhase)

	exp, err := swan.NewExperiment("MemcachedExperiment", phases, "/tmp")
	if err != nil {
		return
	}
	fmt.Println("Starting Memcached Experiment")
	exp.Run()
	fmt.Println("Done")
}
