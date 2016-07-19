package isolation

import (
	"fmt"

	"strconv"
	"strings"
)

// TasksetDecorator defines taskset decorator configuration.
type TasksetDecorator struct {
	cpus []int
}

// NewTasksetDecorator is a constructor for TasksetDecorator object.
func NewTasksetDecorator(cpus []int) TasksetDecorator {
	return TasksetDecorator{
		cpus: cpus,
	}
}

// Decorate implements Decorator interface.
func (t *TasksetDecorator) Decorate(command string) string {
	var cpuList string

	var cpusStr []string
	for _, value := range t.cpus {
		if value >= 0 {
			cpusStr = append(cpusStr, strconv.Itoa(value))
		}
	}

	if len(cpusStr) > 0 {
		cpuList = strings.Join(cpusStr, ",")
	} else {
		cpuList = "0"
	}

	return fmt.Sprintf("taskset -c %s %s", cpuList, command)
}
