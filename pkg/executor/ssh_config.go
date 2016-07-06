package executor

import (
	"io/ioutil"
	"os"
	"os/user"
	"regexp"

	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
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
		return nil, errors.Wrapf(err, "Reading key '%s' failed", keyPath)
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, errors.Wrapf(err, "Parsing private key '%s' failed", keyPath)
	}

	return ssh.PublicKeys(key), nil
}

// ValidateSSHConfig checks if we are able to do remote connection using given host and user.
// Return error if there is blocker (e.g host is not authorized).
func ValidateSSHConfig(host string, user *user.User) error {
	if _, err := os.Stat(user.HomeDir + defaultSSHKeyPath); os.IsNotExist(err) {
		return errors.Errorf("SSH keys not found in '%s'", user.HomeDir+defaultSSHKeyPath)
	}

	// Check if host is self-authorized. If it's localhost we need to grab real hostname.
	if host == "127.0.0.1" || host == "localhost" {
		var err error
		host, err = os.Hostname()
		if err != nil {
			return errors.Wrap(err, "Cannot figure out if localhost is self-authorized")
		}
	}

	authorizedHostsFile, err := os.Open(user.HomeDir + "/.ssh/authorized_keys")
	if err != nil {
		return errors.Wrap(err, "Cannot figure out if localhost is self-authorized")
	}
	authorizedHosts, err := ioutil.ReadAll(authorizedHostsFile)
	if err != nil {
		return errors.Wrapf(err, "Cannot figure out if '%s' is authorized", host)
	}

	re := regexp.MustCompile(host)
	match := re.Find(authorizedHosts)

	if match == nil {
		return errors.Errorf("'%s' is not authorized", host)
	}

	return nil
}

// NewSSHConfig creates a new ssh config for user.
// NOTE: Assumed that private key & authorized host is available in default dirs (<home_dir>/.ssh/).
func NewSSHConfig(host string, port int, user *user.User) (*SSHConfig, error) {
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

func (s SSHConfig) String() string {
	return fmt.Sprintf("")
}
