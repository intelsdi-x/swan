package executor

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
)

func getBinaryNameFromCommand(command string) (string, error) {
	_, name := path.Split(command)
	nameSplit := strings.Split(name, " ")
	if len(nameSplit) == 0 {
		return "", errors.Errorf("Failed to extract command name from '%s'", command)
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
	directoryPrivileges := os.FileMode(0755)

	pwd, err := os.Getwd()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to get working directory")
	}
	outputDir, err := ioutil.TempDir(pwd, prefix+"_"+commandName+"_")
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Failed to create output directory for '%s'\n", commandName)
	}
	if err = os.Chmod(outputDir, directoryPrivileges); err != nil {
		return nil, nil, errors.Wrapf(err, "Failed to set privileges for dir '%s'\n", outputDir)
	}

	filePrivileges := os.FileMode(0644)

	stdoutFileName := path.Join(outputDir, "stdout")
	stdout, err = os.Create(stdoutFileName)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Creating '%s' failed", stdoutFileName)
	}
	if err = stdout.Chmod(filePrivileges); err != nil {
		return nil, nil, errors.Wrapf(err, "Failed to set privileges for file '%s'\n", stdout.Name())
	}

	stderrFileName := path.Join(outputDir, "stderr")
	stderr, err = os.Create(stderrFileName)
	if err != nil {
		os.Remove(stdoutFileName)
		return nil, nil, errors.Wrapf(err, "os.Create failed for path '%s'\n", stderrFileName)
	}
	if err = stderr.Chmod(filePrivileges); err != nil {
		return nil, nil, errors.Wrapf(err, "Failed to set privileges for file '%s'", stderr.Name())
	}

	return stdout, stderr, err
}
