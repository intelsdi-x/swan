package isolation

import (
	"errors"
	"strings"
	"syscall"
)

type namespace struct {
	ns         int
	nsToOption map[int]string
}

// NewNamespace creates new instance of namespace isolation; see: man 7 namespaces.
func NewNamespace(mask int) (Decorator, error) {
	if mask&(syscall.CLONE_NEWIPC|syscall.CLONE_NEWNET|syscall.CLONE_NEWNS|syscall.CLONE_NEWPID|syscall.CLONE_NEWUSER|syscall.CLONE_NEWUTS) == 0 {
		return nil, errors.New("Invalid namespace mask")
	}
	optionMap := make(map[int]string)
	optionMap[syscall.CLONE_NEWPID] = "--fork --pid --mount-proc"
	optionMap[syscall.CLONE_NEWIPC] = "--ipc"
	optionMap[syscall.CLONE_NEWNS] = "--mount"
	optionMap[syscall.CLONE_NEWUTS] = "--uts"
	optionMap[syscall.CLONE_NEWNET] = "--net"
	optionMap[syscall.CLONE_NEWUSER] = "--user"
	return &namespace{ns: mask, nsToOption: optionMap}, nil
}

// Decorate implements Decorator.
func (n *namespace) Decorate(command string) (decorated string) {
	var namespaces []string

	for namespace, option := range n.nsToOption {
		if n.ns&namespace == namespace {
			namespaces = append(namespaces, option)
		}
	}

	return "unshare " + strings.Join(namespaces, " ") + " " + command
}
