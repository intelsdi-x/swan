package main

import (
	"os"

	"github.com/intelsdi-x/snap/control/plugin"
	session "github.com/intelsdi-x/swan/plugin/snap-processor-session-tagging/session-processor"
)

func main() {
	meta := session.Meta()
	plugin.Start(meta, new(session.SessionProcessor), os.Args[1])
}
