package executor

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func getBinaryNameFromCommand(command string) (string, error) {
	_, name := path.Split(command)
	nameSplit := strings.Split(name, " ")
	if len(nameSplit) == 0 {
		return "", fmt.Errorf("Failed to extract command name from %s", command)
	}
	return nameSplit[0], nil
}

func createExecutorOutputFiles(command, prefix string) (stdout, stderr *os.File, err error) {
	if len(command) == 0 {
		return nil, nil, errors.New("Empty command string")
	}

	commandName, err := getBinaryNameFromCommand(command)
	if err != nil {
		return nil, nil, err
	}

	pwd, err := os.Getwd()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to get working directory. Error: %s\n", err.Error())
	}
	outputDir, err := ioutil.TempDir(pwd, prefix+"_"+commandName+"_")
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create output directory for %s. Error: %s\n", commandName,
			err.Error())
	}

	stdoutFileName := path.Join(outputDir, "stdout")
	stdout, err = os.Create(stdoutFileName)
	if err != nil {
		return nil, nil, err
	}

	stderr, err = os.Create(path.Join(outputDir, "stderr"))
	if err != nil {
		os.Remove(stdoutFileName)
		return nil, nil, err
	}

	return stdout, stderr, err
}
