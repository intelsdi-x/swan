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
	task    executor.TaskHandle
	apiPort int
}

// NewSnapd constructs Snapd.
// NOTE(bp): Since Convey test like to overlap tests it is crucial to run snapd on
// different ports.
func NewSnapd(apiPort int) *Snapd {
	return &Snapd{apiPort: apiPort}
}

// Execute starts Snap daemon.
func (s *Snapd) Execute() error {
	l := executor.NewLocal()
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return errors.New("Cannot find GOPATH")
	}

	snapRoot := path.Join(gopath, "src", "github.com", "intelsdi-x", "snap", "build", "bin", "snapd")
	snapCommand := fmt.Sprintf("%s -t 0 -p %d", snapRoot, s.apiPort)

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
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", s.apiPort))
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		defer conn.Close()
		connected = true
	}

	return connected
}
