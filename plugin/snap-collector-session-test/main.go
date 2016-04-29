package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/control/plugin"
	session "github.com/intelsdi-x/swan/plugin/snap-collector-session-test/session-collector"
	"os"
)

func main() {
	logger := log.New()
	
	meta := session.Meta()
	meta.RPCType = plugin.JSONRPC

	logger.Println("starting test snap session collector")

	// Start a collector
	plugin.Start(meta, new(session.SessionCollector), os.Args[1])

	log.Debug("ended test snap session collector")
}
