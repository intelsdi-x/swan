package executor

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Sirupsen/logrus"
)

type taskHandleStopper struct {
	taskHandles []TaskHandle
	sync.Mutex
}

var globalTaskHandleStopper *taskHandleStopper

// RegisterInterruptHandle waits for Interrupt signal and stops unconditionally all taskHandles.
func RegisterInterruptHandle() func() {
	globalTaskHandleStopper = &taskHandleStopper{}
	return globalTaskHandleStopper.registerInterruptHandle()
}

func register(t TaskHandle) {
	globalTaskHandleStopper.register(t)
}

func (ths *taskHandleStopper) registerInterruptHandle() func() {
	ths.Lock()
	defer ths.Unlock()
	logrus.Debugf("clean: interupt hanndle initialized")

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
	ths.Lock()
	defer ths.Unlock()
	// Stop in reverse order.
	if len(ths.taskHandles) > 0 {
		for i := len(ths.taskHandles); i >= 0; i-- {
			taskHandle := ths.taskHandles[i-1]
			logrus.Debugf("clean: stopping '%v'...", taskHandle)
			logrus.Debugf("clean: taskHandle '%v' Stop() returned '%v'", taskHandle, taskHandle.Stop())
		}
	}
}

func (ths *taskHandleStopper) register(t TaskHandle) {
	ths.Lock()
	defer ths.Unlock()
	if ths.taskHandles != nil {
		ths.taskHandles = append(ths.taskHandles, t)
	}
}
