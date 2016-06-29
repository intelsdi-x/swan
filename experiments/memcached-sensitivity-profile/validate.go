package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/utils/sysctl"
)

func validate() error {
	logrus.SetLevel(conf.LogLevel())

	// Warn user about potential issue with SYN flooding of victim machine.
	value, err := sysctl.Get("net.ipv4.tcp_syncookies")
	if err != nil {
		logrus.Debug("Could not read net.ipv4.tcp_syncookies sysctl key: " + err.Error())
	} else if value == "1" {
		logrus.Warn("net.ipv4.tcp_syncookies is enabled on the memcached target and may lead to SYN flooding detection closing mutilate connections.")
	}
	logrus.Debug("Value was %s", value)

	return nil
}
