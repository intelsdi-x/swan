package executor

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/pkg/errors"
)

func getBinaryNameFromCommand(command string) (string, error) {
	argsSplit := strings.Split(command, " ")
	if len(argsSplit) == 0 {
		return "", errors.Errorf("failed to extract command name from %q", command)
	}
	_, name := path.Split(argsSplit[0])
	return name, nil
}

func createExecutorOutputFiles(command, prefix string) (stdout, stderr *os.File, err error) {
	if len(command) == 0 {
		return nil, nil, errors.New("empty command string")
	}

	commandName, err := getBinaryNameFromCommand(command)
	if err != nil {
		return nil, nil, err
	}
	directoryPrivileges := os.FileMode(0755)

	pwd, err := os.Getwd()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get working directory")
	}
	outputDir, err := ioutil.TempDir(pwd, prefix+"_"+commandName+"_")
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to create output directory for %q", commandName)
	}
	if err = os.Chmod(outputDir, directoryPrivileges); err != nil {
		return nil, nil, errors.Wrapf(err, "failed to set privileges for dir %q", outputDir)
	}

	filePrivileges := os.FileMode(0644)

	stdoutFileName := path.Join(outputDir, "stdout")
	stdout, err = os.Create(stdoutFileName)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "creating %q failed", stdoutFileName)
	}
	if err = stdout.Chmod(filePrivileges); err != nil {
		return nil, nil, errors.Wrapf(err, "failed to set privileges for file %q", stdout.Name())
	}

	stderrFileName := path.Join(outputDir, "stderr")
	stderr, err = os.Create(stderrFileName)
	if err != nil {
		os.Remove(stdoutFileName)
		return nil, nil, errors.Wrapf(err, "os.Create failed for path %q", stderrFileName)
	}
	if err = stderr.Chmod(filePrivileges); err != nil {
		return nil, nil, errors.Wrapf(err, "failed to set privileges for file %q", stderr.Name())
	}

	return stdout, stderr, err
}

func readTail(filePath string) (tail string, err error) {
	output, err := exec.Command("tail", filePath).CombinedOutput()

	if err != nil {
		return "", errors.Wrapf(err, "could not read tail of %q", filePath)
	}

	return string(output), nil
}
