package main

import (
	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/swan/misc/snap-plugin-collector-mutilate/mutilate"
	"os"
	"time"
)

func main() {
	meta := mutilate.Meta()
	meta.RPCType = plugin.JSONRPC

	plugin.Start(meta, mutilate.NewMutilate(time.Now()), os.Args[1])
}
