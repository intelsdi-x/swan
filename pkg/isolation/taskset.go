package isolation

import (
	"io/ioutil"
	"os/exec"
	"path"
	"strconv"
    "strings"
    "fmt"

	"github.com/pkg/errors"
)

// Taskset defines input data
type TasksetDecorator struct {
	cpus []int
}
 

// NewTasksetDecorator is a constructor for TasksetDecorator object. 
func NewTasksetDecorator(cpus []int) Numa {
    return TasksetDecorator{
        cpus: cpus,
    }
}

// Decorate implements Decorator interface.
func (t *TasksetDecorator) Decorate(command string) string {
    cpusStr := make([]string, len(t.cpus))
    for idx, value := range t.cpus {
        cpusStr[idx] := strconv.Itoa(value)
    }
	return fmt.Sprintf("taskset -c %s %s", strings.Join(cpusStr), command)
}
