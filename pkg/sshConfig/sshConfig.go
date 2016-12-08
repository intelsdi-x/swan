package sshConfig

import (
	"github.com/hypersleep/easyssh"
)

func NewSshConfig(user string, server string, keyPath string, port string) *easyssh.MakeConfig {
        return &easyssh.MakeConfig{
                user,
                server,
                keyPath,
                port,
        }
}
