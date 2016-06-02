package cgroup

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	pth "path"
	"strconv"
	"strings"
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/isolation"
)

const (
	// DefaultCommandTimeout is the default amount of time to wait for
	// dispatched commands to finish executing.
	DefaultCommandTimeout = 1 * time.Second
)

// Cgroup represents a Linux control group.
// See https://www.kernel.org/doc/Documentation/cgroup-v1/cgroups.txt
//
// Usage of this interface requires the libcgroup tools to be installed
// on the system. This library interacts with cgroups by shelling out to
// utility programs like `cgcreate`, `cgexec`, `cgget` and friends.
type Cgroup interface {
	isolation.Isolation

	// Path returns this cgroup's controllers.
	Controllers() []string

	// Path returns this cgroup's path in the hierarchy.
	Path() string

	// AbsPath returns the absolute path to this cgroup within the
	// VFS mount for the specified controller.
	// Returns the empty string if the controller is not a member of
	// this cgroup's controllers.
	// e.g. `d, err := cgroup.AbsPath("cpuset")`
	AbsPath(controller string) string

	// Exists returns `true` iff this cgroup is present in all of its
	// controller hierarchies.
	Exists() (bool, error)

	// Destroy removes this cgroup.
	// If recursive is specified, also destroy this cgroup's children.
	// Returns an error if this cgroup has children but recurvisve was not
	// specified.
	// Returns an error if this cgroup cannot be destroyed.
	Destroy(recursive bool) error

	// Parent returns the direct ancestor of this cgroup, or nil if this
	// is the root.
	Parent() Cgroup

	// Tasks returns the pids for this cgroup and the supplied controller.
	// Returns an error if the controller is not a member of
	// this cgroup's controllers.
	//
	// NB: Linux pid range is [0,  2^22]; see /proc/sys/kernel/pid_max.
	Tasks(controller string) (isolation.IntSet, error)

	// Get returns the value of an attribute for this Cgroup.
	Get(name string) (string, error)

	// Set overwrites the value of an attribute for this Cgroup.
	Set(name string, value string) error
}

// NewCgroup returns a new Cgroup with the supplied controllers and path.
// Returns an error if no controllers are specified or the path is empty.
func NewCgroup(controllers []string, path string) (Cgroup, error) {
	return NewCgroupWithExecutor(controllers,
		path,
		executor.NewLocal(),
		DefaultCommandTimeout)
}

// NewCgroupWithExecutor returns a new Cgroup with the supplied controllers,
// path and executor. Returns an error if no controllers are specified or
// the path is empty.
func NewCgroupWithExecutor(controllers []string,
	path string,
	executor executor.Executor,
	cmdTimeout time.Duration) (Cgroup, error) {
	if len(controllers) == 0 {
		return nil, fmt.Errorf("No controllers specified for cgroup")
	}
	if path == "" {
		return nil, fmt.Errorf("Empty path specified for cgroup")
	}
	if executor == nil {
		return nil, fmt.Errorf("Nil executor supplied for cgroup")
	}
	canonicalPath := pth.Join("/", path)
	return &cgroup{controllers, canonicalPath, executor, cmdTimeout}, nil
}

// The cgroup struct implements the Cgroup interface.
type cgroup struct {
	controllers []string
	path        string
	executor    executor.Executor
	cmdTimeout  time.Duration
}

func (cg *cgroup) Controllers() []string {
	return cg.controllers
}

func (cg *cgroup) Path() string {
	return cg.path
}

func (cg *cgroup) AbsPath(controller string) string {
	p, err := SubsysPath(controller, cg.executor, cg.cmdTimeout)
	if err != nil {
		return ""
	}
	return pth.Join(p, cg.path)
}

func (cg *cgroup) Exists() (bool, error) {
	for _, ctrl := range cg.controllers {
		out, err := cg.cmdOutput("lscgroup", "-g", fmt.Sprintf("%s:%s", ctrl, cg.path))
		if err != nil {
			return false, err
		}
		if strings.Count(string(out), "\n") < 1 {
			return false, nil
		}
	}
	return true, nil
}

func (cg *cgroup) Create() error {
	_, err := cg.cmdOutput("cgcreate", "-g", cg.spec())
	return err
}

func (cg *cgroup) Destroy(recursive bool) error {
	if recursive {
		_, err := cg.cmdOutput("cgdelete", "--recursive", "-g", cg.spec())
		return err
	}
	_, err := cg.cmdOutput("cgdelete", "-g", cg.spec())
	return err
}

func (cg *cgroup) Parent() Cgroup {
	if cg.path == "/" {
		return nil
	}
	parentPath, _ := pth.Split(cg.path)
	// Discarding errors here because controllers and path are both
	// guaranteed to be non-empty.
	p, _ := NewCgroup(cg.controllers, parentPath)
	return p
}

func (cg *cgroup) Tasks(controller string) (isolation.IntSet, error) {
	d := cg.AbsPath(controller)
	if d == "" {
		return nil, fmt.Errorf("Failed to read absolute path for controller %s", controller)
	}

	tf, err := os.Open(pth.Join(d, "tasks"))
	defer tf.Close()
	if err != nil {
		return nil, err
	}

	pids := isolation.NewIntSet()
	s := bufio.NewScanner(tf)
	for s.Scan() {
		t, err := strconv.Atoi(s.Text())
		if err != nil {
			return nil, err
		}
		pids.Add(t)
	}

	// After Scan returns false, the Err method returns any scanning
	// operations, except in case of EOF.
	if s.Err() != nil {
		return nil, s.Err()
	}

	return pids, err
}

func (cg *cgroup) Get(name string) (string, error) {
	out, err := cg.cmdOutput("cgget", "-nv", "--variable", name, cg.path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (cg *cgroup) Set(name string, value string) error {
	_, err := cg.cmdOutput("cgset", "-r", fmt.Sprintf("%s=%s", name, value), cg.path)
	if err != nil {
		return err
	}
	return nil
}

func (cg *cgroup) Clean() error {
	return cg.Destroy(true)
}

func (cg *cgroup) Decorate(command string) string {
	return fmt.Sprintf("cgexec -g %s %s", cg.spec(), command)
}

func (cg *cgroup) Isolate(PID int) error {
	_, err := cg.cmdOutput("cgclassify", "-g", cg.spec(), strconv.Itoa(PID))
	return err
}

// Internal helpers for getting command output.
func cmdOutput(executor executor.Executor, cmdTimeout time.Duration, argv ...string) (string, error) {
	cmd := strings.Join(argv, " ")
	task, err := executor.Execute(cmd)
	defer task.EraseOutput()
	if err != nil {
		return "", err
	}
	if ok := task.Wait(cmdTimeout); !ok {
		return "", fmt.Errorf("Timed out waiting for command: %s", cmd)
	}
	code, err := task.ExitCode()
	if err != nil {
		return "", err
	}
	if code != 0 {
		return "", fmt.Errorf("Command exited with code %d: %s", code, cmd)
	}
	oFile, err := task.StdoutFile()
	if err != nil {
		return "", err
	}
	bytes, err := ioutil.ReadFile(oFile.Name()) // assume small output
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (cg *cgroup) cmdOutput(argv ...string) (string, error) {
	return cmdOutput(cg.executor, cg.cmdTimeout, argv...)
}

// Internal helper for creating libcgroup-tools compatible args.
// Returns a string like 'cpu,cpuset:/my/cool/group'.
func (cg *cgroup) spec() string {
	return fmt.Sprintf("%s:%s", strings.Join(cg.controllers, ","), cg.path)
}
