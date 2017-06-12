package executor

import (
	"sync"
	"time"
)

// ChainedTaskHandle is an links Launchers in a way that
// one will be launched after another.
type ChainedTaskHandle struct {
	TaskHandle

	chainedLaunchers []Launcher

	encounteredError error

	chainFinished chan struct{}
	stopChain     chan struct{}
	stopOnce      sync.Once
}

// NewChainedTaskHandle returns TaskHandle that executes current handle, and will
// launch Launcher when handle will finish it's execution.
func NewChainedTaskHandle(handle TaskHandle, launcher ...Launcher) TaskHandle {
	launchers := make([]Launcher, 0)
	for _, l := range launcher {
		launchers = append(launchers, l)
	}

	chained := ChainedTaskHandle{
		TaskHandle:       handle,
		chainedLaunchers: launchers,
		chainFinished:    make(chan struct{}),
		stopChain:        make(chan struct{}, 1),
	}

	go chained.watch()

	return &chained
}

func (cth *ChainedTaskHandle) watch() {
	waitChan := getWaitChannel(cth.TaskHandle)
	select {
	case err := <-waitChan:
		if err != nil {
			cth.encounteredError = err
			close(cth.chainFinished)
			return
		}
	case <-cth.stopChain:
		err := cth.TaskHandle.Stop()
		cth.encounteredError = err
		close(cth.chainFinished)
		return
	}

	// Run all chained Launchers.
	for _, launcher := range cth.chainedLaunchers {
		chainedHandle, err := launcher.Launch()
		if err != nil {
			cth.encounteredError = err
			close(cth.chainFinished)
			return
		}

		waitChan = getWaitChannel(chainedHandle)
		select {
		case err := <-waitChan:
			cth.encounteredError = err
			if err != nil {
				cth.encounteredError = err
				close(cth.chainFinished)
				return
			}
		case <-cth.stopChain:
			err := chainedHandle.Stop()
			cth.encounteredError = err
			close(cth.chainFinished)
			return
		}
	}

	close(cth.chainFinished)
	return
}

// Stop stops all execution of ChainedTaskHandle.
func (cth *ChainedTaskHandle) Stop() error {
	cth.stopOnce.Do(func() {
		cth.stopChain <- struct{}{}
	})

	// Wait for chain to stop.
	_, err := cth.Wait(0)
	return err
}

// Wait waits for all tasks in ChainedTaskHandle to finish.
func (cth *ChainedTaskHandle) Wait(timeout time.Duration) (bool, error) {
	timeoutChannel := getWaitTimeoutChan(timeout)

	select {
	case <-timeoutChannel:
		return false, nil
	case <-cth.chainFinished:
		return true, cth.encounteredError
	}
}

// Status returns current TaskState.
func (cth *ChainedTaskHandle) Status() TaskState {
	select {
	case <-cth.chainFinished:
		return TERMINATED
	default:
		return RUNNING
	}
}
