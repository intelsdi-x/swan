package main

import (
	"os"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/swan/misc/snap-plugin-collector-mutilate/mutilate"
)

func main() {
	meta := mutilate.Meta()
	plugin.Start(meta, mutilate.NewMutilate(time.Now()), os.Args[1])
}
