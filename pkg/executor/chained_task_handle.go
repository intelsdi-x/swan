package executor

import (
	"sync"
	"time"
)

type ChainedTaskHandle struct {
	TaskHandle

	chainedLauncher Launcher

	encounteredError error

	chainFinished chan struct{}
	stopChain     chan struct{}
	stopOnce      sync.Once
}

func NewChainedTaskHandle(handle TaskHandle, launcher Launcher) TaskHandle {
	chained := ChainedTaskHandle{
		TaskHandle:      handle,
		chainedLauncher: launcher,
		chainFinished:   make(chan struct{}),
		stopChain:       make(chan struct{}, 1),
	}

	go chained.watcher()

	return &chained
}

func (cth *ChainedTaskHandle) watcher() {
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
	cth.stopOnce.Do(func() {
		cth.stopChain <- struct{}{}
	})

	// Wait for chain to stop.
	_, err := cth.Wait(0)
	return err
}

func (cth *ChainedTaskHandle) Wait(timeout time.Duration) (bool, error) {
	timeoutChannel := getWaitTimeoutChan(timeout)

	select {
	case <-timeoutChannel:
		return false, nil
	case <-cth.chainFinished:
		return true, cth.encounteredError
	}
}

func (cth *ChainedTaskHandle) Status() TaskState {
	select {
	case <-cth.chainFinished:
		return TERMINATED
	default:
		return RUNNING
	}
}
