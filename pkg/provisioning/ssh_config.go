package provisioning

import (
	"io/ioutil"
	"golang.org/x/crypto/ssh"
)

// SSHConfig with clientConfig, host and port to connect.
type SSHConfig struct {
	clientConfig *ssh.ClientConfig
	host   string
	port   int
}

// NewsshConfig creates a new ssh config.
func NewsshConfig(clientConfig *ssh.ClientConfig, host string, port int) *SSHConfig {
	return &SSHConfig{
		clientConfig,
		host,
		port,
	}
}

// get AuthMethod which uses given key.
func publicKeyFile(keyPath string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(keyPath)
	if err != nil {
		panic(err)
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		panic(err)
		return nil
	}

	return ssh.PublicKeys(key)
}

// NewClientConfig create client config with credentials for ssh connection.
func NewClientConfig(username string, keyPath string) *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			publicKeyFile(keyPath),
		},
	}
}
