package testhelpers

import (
	"errors"
	"fmt"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/athena/pkg/executor"
	"math/rand"
	"os"
	"path"
	"time"
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

// Start starts Snap daemon.
func (s *Snapd) Start() error {
	l := executor.NewLocal()
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return errors.New("Cannot find GOPATH")
	}

	snapRoot := path.Join(
		gopath, "src", "github.com", "intelsdi-x", "snap", "build", "bin", "snapd")
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
	retries := 100
	isConnected := false
	cli, err := client.New(fmt.Sprintf("http://127.0.0.1:%d", s.apiPort), "v1", true)
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
