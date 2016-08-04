package fs

import (
	"fmt"
	"os/exec"

	"github.com/pkg/errors"
)

func ReadTail(filePath string, lineCount int) (tail string, err error) {
	lineCountParam := fmt.Sprintf("-n %d", lineCount)
	output, err := exec.Command("tail", lineCountParam, filePath).CombinedOutput()

	if err != nil {
		return "", errors.Wrapf(err, "could not read tail of %q", filePath)
	}

	return string(output), nil
}
