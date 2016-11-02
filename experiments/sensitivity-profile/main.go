package main

import (
	"github.com/intelsdi-x/athena/pkg/utils/errutil"
	"github.com/intelsdi-x/swan/experiments/sensitivity-profile/common"
)

func main() {
	err := common.RunExperiment()
	errutil.Check(err)
}
