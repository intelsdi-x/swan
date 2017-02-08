package main

import (
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/intelsdi-x/swan/misc/snap-plugin-collector-mutilate/mutilate"
)

func main() {
	plugin.StartCollector(mutilate.NewMutilate(time.Now()), mutilate.NAME, mutilate.VERSION, plugin.CacheTTL(1*time.Second))
}
