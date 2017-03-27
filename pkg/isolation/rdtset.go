package isolation

import (
	"fmt"
	"os/exec"

	"bytes"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

// Rdtset is an instance of Decorator that used rdtset command for isolation. It allows to set CPU affinity and allocate cache available to those CPUs.
// See documentation at: experiments/memcached-cat/README.md
type Rdtset struct {
	CPURange string
	Mask     int
}

// Decorate implements Decorator interface
func (r Rdtset) Decorate(command string) (decorated string) {
	decorated = fmt.Sprintf("rdtset.sh -v -c %s -t 'l3=%#x;cpu=%s' %s", r.CPURange, r.Mask, r.CPURange, command)
	logrus.Debugf("Command decorated with rdtset: %s", decorated)

	return
}

// CleanRDTAssingments cleans any existing RDT RMID's assignment.
func CleanRDTAssingments() (string, error) {
	cmd := exec.Command("pqos.sh", "-R")
	outputtedBytes, err := cmd.CombinedOutput()
	buf := bytes.NewBuffer(outputtedBytes)
	output := buf.String()
	if err != nil {
		return output, errors.Wrapf(err, "pqos -R failed. Output: %q", output)
	}

	return output, nil
}
