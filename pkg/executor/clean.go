package executor

import (
	"container/list"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Sirupsen/logrus"
)

type taskHandleStopper struct {
	taskHandles *list.List
	mutex       sync.Mutex
}

var globalTaskHandleStopper taskHandleStopper

// RegisterInterruptHandle waits for Interrupt signal and stops unconditionally all taskHandles.
// Additionall returns a function that can be used in defer to handle panics in main function.
func RegisterInterruptHandle() func() {
	return globalTaskHandleStopper.registerInterruptHandle()
}

func register(t TaskHandle) {
	globalTaskHandleStopper.register(t)
}

func unregister(t TaskHandle) {
	globalTaskHandleStopper.unregister(t)
}

func (ths *taskHandleStopper) registerInterruptHandle() func() {
	ths.mutex.Lock()
	defer ths.mutex.Unlock()
	logrus.Debugf("clean: interupt hanndle initialized")
	ths.taskHandles = list.New()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		logrus.Debugf("clean: stopAllTaskHandles on signal '%v'", <-c)
		ths.stopAllTaskHandles()
		os.Exit(1)
	}()
	return ths.stopAllTaskHandles
}

func (ths *taskHandleStopper) stopAllTaskHandles() {
	ths.mutex.Lock()
	defer ths.mutex.Unlock()
	if ths.taskHandles != nil {
		// Stop in reverse order.
		for e := ths.taskHandles.Back(); e != nil; e = e.Prev() {
			taskHandle := e.Value.(TaskHandle)
			logrus.Debugf("clean: taskHandle '%v' Stop() returned '%v'", taskHandle, taskHandle.Stop())
		}
	}
}

func (ths *taskHandleStopper) register(t TaskHandle) {
	ths.mutex.Lock()
	defer ths.mutex.Unlock()
	if ths.taskHandles != nil {
		ths.taskHandles.PushBack(t)
	}
}

func (ths *taskHandleStopper) unregister(t TaskHandle) {
	ths.mutex.Lock()
	defer ths.mutex.Unlock()
	if ths.taskHandles != nil {
		for e := ths.taskHandles.Front(); e != nil; e = e.Next() {
			taskHandle := e.Value.(TaskHandle)
			if t == taskHandle {
				ths.taskHandles.Remove(e)
			}
		}
	}
}
