package executor

import (
	"io/ioutil"

	"errors"
	"fmt"
	"github.com/intelsdi-x/swan/pkg/net"
	"golang.org/x/crypto/ssh"
	"os"
	"os/user"
	"regexp"
)

const (
	// DefaultSSHPort represent default port of SSH server (22).
	DefaultSSHPort    = 22
	defaultSSHKeyPath = "/.ssh/id_rsa"
)

// SSHConfig with clientConfig, host and port to connect.
type SSHConfig struct {
	ClientConfig *ssh.ClientConfig
	Host         string
	Port         int
}

// getAuthMethod which uses given key.
func getAuthMethod(keyPath string) (ssh.AuthMethod, error) {
	buffer, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, err
	}

	return ssh.PublicKeys(key), nil
}

// validateConfig checks if we are able to do remote connection using given host and user.
// Return error if there is blocker (e.g host is not authorized).
func validateConfig(host string, user *user.User) error {
	if _, err := os.Stat(user.HomeDir + defaultSSHKeyPath); os.IsNotExist(err) {
		return fmt.Errorf("SSH keys not found in %s", user.HomeDir+defaultSSHKeyPath)
	}

	// Check if host is self-authorized. If it's localhost we need to grab real hostname.
	if net.IsAddrLocal(host) {
		var err error
		host, err = os.Hostname()
		if err != nil {
			return errors.New("Cannot figure out if localhost is self-authorized")
		}

		// TODO(bp): [SCE-423] Make this for remote hosts as well, when we have /etc/hosts
		// propagated on our hosts.
		// Currently we don't have /etc/hosts propagated and ssh cannot resolve the host.
		// Even if we authorize the IP of remote machine it is saved using hostname in
		// authorized keys file.
		authorizedHostsFile, err := os.Open(user.HomeDir + "/.ssh/authorized_keys")
		if err != nil {
			return errors.New("Cannot figure out if localhost is self-authorized: " + err.Error())
		}
		authorizedHosts, err := ioutil.ReadAll(authorizedHostsFile)
		if err != nil {
			return fmt.Errorf("Cannot figure out if %s is authorized: %s", host, err.Error())
		}

		re := regexp.MustCompile(host)
		match := re.Find(authorizedHosts)

		if match == nil {
			return fmt.Errorf("%s is not authorized", host)
		}
	}

	return nil
}

// NewSSHConfig creates a new ssh config for user.
// NOTE: Assumed that private key & authorized host is available in default dirs (<home_dir>/.ssh/).
func NewSSHConfig(host string, port int, user *user.User) (*SSHConfig, error) {
	err := validateConfig(host, user)
	if err != nil {
		return nil, err
	}

	authMethod, err := getAuthMethod(user.HomeDir + defaultSSHKeyPath)
	if err != nil {
		return nil, err
	}

	clientConfig := &ssh.ClientConfig{
		User: user.Username,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
	}

	return &SSHConfig{
		ClientConfig: clientConfig,
		Host:         host,
		Port:         port,
	}, nil
}
