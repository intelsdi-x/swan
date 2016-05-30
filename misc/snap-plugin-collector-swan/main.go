package main

import (
	"github.com/intelsdi-x/snap/control/plugin"
	collector "github.com/intelsdi-x/swan/misc/snap-plugin-collector-swan/swan-metrics-collector"
	"os"
)

func main() {
	meta := collector.Meta()
	meta.RPCType = plugin.JSONRPC

	// Start server
	go collector.StartServer()

	// Start a collector.
	plugin.Start(meta, new(collector.SessionCollector), os.Args[1])
}
