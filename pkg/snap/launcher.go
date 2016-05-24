package snap

import (
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
)

// SessionHandle is handle for Snap Collection session. It can be stopped from here.
// NOTE: In SnapSessionHandle Stop() method needs to ensure that the session has completed it's work.
// We can move that to generic collection in future - for now we only use snap.
type SessionHandle interface {
	IsRunning() bool
	StopAndRemove() error
	WaitToBeHitOnce()
}

// SessionLauncher starts Snap Collection session and returns handle to that session.
type SessionLauncher interface {
	LaunchSession(executor.TaskInfo, phase.Session) (SessionHandle, error)
}
