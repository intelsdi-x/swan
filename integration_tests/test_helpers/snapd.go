package testhelpers

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/pkg/errors"
)

// Snapd represents Snap daemon used in tests.
type Snapd struct {
	task    executor.TaskHandle
	apiPort int
}

// NewSnapd constructs Snapd on random high port.
func NewSnapd() *Snapd {
	randomHighPort := rand.Intn(32768-10000) + 10000
	return NewSnapdOnPort(randomHighPort)
}

// NewSnapdOnPort constructs Snapd on chosen port.
func NewSnapdOnPort(apiPort int) *Snapd {
	return &Snapd{apiPort: apiPort}
}

// Start starts Snap daemon and wait until it is responsive.
func (s *Snapd) Start() error {
	l := executor.NewLocal()
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return errors.New("Cannot find GOPATH")
	}
	snapdPath := path.Join(gopath, "bin", "snapd")
	snapCommand := fmt.Sprintf("%s -t 0 -p %d", snapdPath, s.apiPort)

	taskHandle, err := l.Execute(snapCommand)
	if err != nil {
		return err
	}

	if !s.Connected() {
		taskHandle.Stop()
		taskHandle.Clean()
		taskHandle.EraseOutput()
		return errors.Errorf("could not connect to snapd on %q", s.getSnapdAddress())
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
	retries := 100
	isConnected := false
	cli, err := client.New(s.getSnapdAddress(), "v1", true)
	if err != nil {
		return isConnected
	}
	for i := 0; i < retries; i++ {
		if cli.GetPlugins(false).Err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		isConnected = true
	}

	return isConnected
}

// Port returns port Snapd is listening.
func (s *Snapd) Port() int {
	return s.apiPort
}

func (s *Snapd) getSnapdAddress() string {
	return fmt.Sprintf("http://127.0.0.1:%d", s.apiPort)
}
