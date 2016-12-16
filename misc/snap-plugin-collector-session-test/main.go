package main

import (
	"os"

	"github.com/intelsdi-x/snap/control/plugin"
	session "github.com/intelsdi-x/swan/misc/snap-plugin-collector-session-test/session-collector"
)

func main() {
	meta := session.Meta()

	// Start a collector.
	plugin.Start(meta, new(session.SessionCollector), os.Args[1])
}
