package executor

import (
	"golang.org/x/crypto/ssh"
	"io/ioutil"
)

// SSHConfig with clientConfig, host and port to connect.
type SSHConfig struct {
	clientConfig *ssh.ClientConfig
	host         string
	port         int
}

// NewSSHConfig creates a new ssh config.
func NewSSHConfig(clientConfig *ssh.ClientConfig, host string, port int) *SSHConfig {
	return &SSHConfig{
		clientConfig,
		host,
		port,
	}
}

// NewClientConfig create client config with credentials for ssh connection.
func NewClientConfig(username string, keyPath string) (*ssh.ClientConfig, error) {
	authMethod, err := publicKeyFile(keyPath)
	if err != nil {
		return nil, err
	}
	return &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
	}, nil
}

// get AuthMethod which uses given key.
func publicKeyFile(keyPath string) (ssh.AuthMethod, error) {
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
