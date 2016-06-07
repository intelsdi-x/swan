package executor

import (
	"io/ioutil"

	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
	"os/user"
	"regexp"
)

const (
	// DefaultSSHPort represent default port of SSH server (22).
	DefaultSSHPort = 22
)

// SSHConfig with clientConfig, host and port to connect.
type SSHConfig struct {
	ClientConfig *ssh.ClientConfig
	Host         string
	Port         int
}

func getDefaultPrivateKeyPath(user *user.User) string {
	return user.HomeDir + "/.ssh/id_rsa"
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
	if _, err := os.Stat(getDefaultPrivateKeyPath(user)); os.IsNotExist(err) {
		return fmt.Errorf("SSH keys not found in %s", getDefaultPrivateKeyPath(user))
	}

	// Check if host is self-authorized. If localhost we need to grab real hostname.
	if host == "127.0.0.1" || host == "localhost" {
		var err error
		host, err = os.Hostname()
		if err != nil {
			return errors.New("Cannot figure out if localhost is self-authorized")
		}
	}

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

	return nil
}

// NewSSHConfig creates a new ssh config for user.
// NOTE: Assumed that private key & authorized host is available in default dirs (<home_dir>/.ssh/).
func NewSSHConfig(host string, port int, user *user.User) (*SSHConfig, error) {
	err := validateConfig(host, user)
	if err != nil {
		return nil, err
	}

	authMethod, err := getAuthMethod(getDefaultPrivateKeyPath(user))
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
