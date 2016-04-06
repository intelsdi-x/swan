package provisioning

import (
	"io/ioutil"
	"golang.org/x/crypto/ssh"
)

// ssh config with clientConfig, host and port to connect
type SshConfig struct {
	clientConfig *ssh.ClientConfig
	host   string
	port   int
}

// Create new ssh config
func NewsshConfig(clientConfig *ssh.ClientConfig, host string, port int) *SshConfig {
	return &SshConfig{
		clientConfig,
		host,
		port,
	}
}

// get AuthMethod which uses given key
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

// Create client config with credentials for ssh connection
func NewClientConfig(username string, keyPath string) *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			publicKeyFile(keyPath),
		},
	}
}
