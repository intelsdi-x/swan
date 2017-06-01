package executor

import (
	"time"
)

type ChainedTaskHandle struct {
	TaskHandle

	chainedLauncher Launcher

	encounteredError error

	chainFinished chan struct{}
	stopChain     chan struct{}
}

func NewChainedTaskHandle(handle TaskHandle, launcher Launcher) TaskHandle {
	var result ChainedTaskHandle

	result.TaskHandle = handle
	result.chainedLauncher = launcher
	result.chainFinished = make(chan struct{})
	result.stopChain = make(chan struct{})

	go result.watcher()

	return result
}

func (cth *ChainedTaskHandle) watcher() {
	// Wait for initial TaskHandle to finish.
	waitChan := GetWaitChannel(cth.TaskHandle)
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
	chainedHandle, err := cth.chainedLauncher.Launch()
	if err != nil {
		cth.encounteredError = err
		close(cth.chainFinished)
		return
	}
	waitChan = GetWaitChannel(chainedHandle)
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

	close(cth.chainFinished)
	return
}

func (cth *ChainedTaskHandle) Stop() error {
	// Try to stop the chain.
	select {
	case cth.stopChain <- struct{}{}:
	default:
	}

	// Wait for chain to stop.
	_, err := cth.Wait(0)
	return err
}

func (cth *ChainedTaskHandle) Wait(timeout time.Duration) (bool, error) {
	select {
	case <-time.After(timeout):
		return false, nil
	case <-cth.chainFinished:
		return true, cth.encounteredError
	}
}
