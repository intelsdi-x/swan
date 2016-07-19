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
	return tasksetDecorator{
		cpus: cpus,
	}
}

// Decorate implements Decorator interface.
func (t *TasksetDecorator) Decorate(command string) string {
	cpusStr := make([]string, len(t.cpus))
	for idx, value := range t.cpus {
		cpusStr[idx] = strconv.Itoa(value)
	}

	return fmt.Sprintf("taskset -c %s %s", strings.Join(cpusStr, " "), command)
}
