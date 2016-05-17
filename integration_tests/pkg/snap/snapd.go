package snap

import (
	"errors"
	"fmt"
	"github.com/intelsdi-x/swan/pkg/executor"
	"net"
	"os"
	"path"
	"time"
)

// Snapd represents Snap daemon used in tests.
type Snapd struct {
	task executor.TaskHandle
}

// NewSnapd constructs Snapd.
func NewSnapd() *Snapd {
	return &Snapd{}
}

// Execute starts Snap daemon.
func (s *Snapd) Execute() error {
	l := executor.NewLocal()
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return errors.New("Cannot find GOPATH")
	}

	snapRoot := path.Join(gopath, "src", "github.com", "intelsdi-x", "snap", "build", "bin", "snapd")
	snapCommand := fmt.Sprintf("%s -t 0", snapRoot)

	taskHandle, err := l.Execute(snapCommand)
	if err != nil {
		return err
	}

	s.task = taskHandle
	return nil
}

// Stop stops Snap daemon.
func (s *Snapd) Stop() error {
	if s.task == nil {
		return errors.New("Snapd not started: cannot find task")
	}

	return s.task.Stop()
}

// CleanAndEraseOutput cleans and removes Output.
func (s *Snapd) CleanAndEraseOutput() error {
	if s.task == nil {
		return errors.New("Snapd not started: cannot find task")
	}

	s.task.Clean()
	return s.task.EraseOutput()
}

// Connected checks if we can connect to Snap daemon.
func (s *Snapd) Connected() bool {
	retries := 5
	connected := false
	for i := 0; i < retries; i++ {
		conn, err := net.Dial("tcp", "127.0.0.1:8181")
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		defer conn.Close()
		connected = true
	}

	return connected
}
