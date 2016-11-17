package main

import (
	"github.com/intelsdi-x/athena/pkg/utils/errutil"
	"github.com/intelsdi-x/swan/experiments/specjbb-sensitivity-profile/common"
)

func main() {
	err := common.RunExperiment()
	errutil.Check(err)
}
