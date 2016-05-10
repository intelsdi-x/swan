package main

import (
	"github.com/intelsdi-x/snap/control/plugin"
	session "github.com/intelsdi-x/swan/misc/snap-plugin-collector-session-test/session-collector"
	"os"
)

func main() {
	meta := session.Meta()
	meta.RPCType = plugin.JSONRPC

	// Start a collector.
	plugin.Start(meta, new(session.SessionCollector), os.Args[1])
}
