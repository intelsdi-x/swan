package executor

import (
	"github.com/intelsdi-x/swan/pkg/net"
	"os/user"
)

// CreateExecutor is factory for executor depending on ip provided. In case of localhost it returns
// LocalExecutor otherwise it returns Remote With default ssh config for current user.
func CreateExecutor(ip string) (Executor, error) {
	// NOTE: We don't want to ssh on localhost if not needed - this enables ease of use inside
	// docker with net=host flag.
	if net.IsAddrLocal(ip) {
		return NewLocal(), nil
	}

	user, err := user.Current()
	if err != nil {
		return nil, err
	}

	sshConfig, err := NewSSHConfig(ip, DefaultSSHPort, user)
	if err != nil {
		return nil, err
	}

	return NewRemote(sshConfig), nil
}
