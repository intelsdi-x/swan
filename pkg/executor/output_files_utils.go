package executor

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/utils/err_collection"
	"github.com/pkg/errors"
)

const outputFilePrivileges = os.FileMode(0644)

func getBinaryNameFromCommand(command string) (string, error) {
	argsSplit := strings.Split(command, " ")
	if len(argsSplit) == 0 {
		return "", errors.Errorf("failed to extract command name from %q", command)
	}
	_, name := path.Split(argsSplit[0])
	return name, nil
}

// createOutputDirectory creates directory for executor output and returns path to it when successful, or error if not.
func createOutputDirectory(command string, prefix string) (createdDirectoryPath string, err error) {
	if len(command) == 0 {
		return "", errors.New("empty command string")
	}

	commandName, err := getBinaryNameFromCommand(command)
	if err != nil {
		return "", err
	}
	directoryPrivileges := os.FileMode(0755)

	pwd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "failed to get working directory")
	}
	createdDirectoryPath, err = ioutil.TempDir(pwd, prefix+"_"+commandName+"_")
	if err != nil {
		return "", errors.Wrapf(err, "failed to create output directory for %q", commandName)
	}
	if err = os.Chmod(createdDirectoryPath, directoryPrivileges); err != nil {
		os.RemoveAll(createdDirectoryPath)
		return "", errors.Wrapf(err, "failed to set privileges for dir %q", createdDirectoryPath)
	}

	return createdDirectoryPath, nil
}

func createExecutorOutputFiles(outputDir string) (stdoutFile, stderrFile *os.File, err error) {
	stdoutFileName := path.Join(outputDir, "stdout")
	stdoutFile, err = os.OpenFile(stdoutFileName, os.O_WRONLY|os.O_SYNC|os.O_EXCL|os.O_CREATE, outputFilePrivileges)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "creating %q failed", stdoutFileName)
	}

	stderrFileName := path.Join(outputDir, "stderr")
	stderrFile, err = os.OpenFile(stderrFileName, os.O_WRONLY|os.O_SYNC|os.O_EXCL|os.O_CREATE, outputFilePrivileges)
	if err != nil {
		// Clean created stdout.
		stdoutFile.Close()
		os.Remove(stdoutFileName)
		return nil, nil, errors.Wrapf(err, "os.Create failed for path %q", stderrFileName)
	}

	return stdoutFile, stderrFile, nil
}

func syncAndClose(file *os.File) error {
	var errCol errcollection.ErrorCollection
	err := file.Sync()
	if err != nil {
		errCol.Add(err)
		log.Errorf("Cannnot sync stdout file: %s", err.Error())
	}
	err = file.Close()
	if err != nil {
		errCol.Add(err)
		log.Errorf("Cannot close stdout file: %s", err.Error())
	}
	return errCol.GetErrIfAny()
}

func openFile(fileName string) (*os.File, error) {
	if _, err := os.Stat(fileName); err != nil {
		return nil, errors.Wrapf(err, "unable to stat file at %q", fileName)
	}

	file, err := os.Open(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to open file at %q", fileName)
	}

	return file, nil
}

// removeDirectory removes directory if exists.
func removeDirectory(directory string) error {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return nil
	}

	if err := os.RemoveAll(directory); err != nil {
		return errors.Wrapf(err, "os.RemoveAll of directory %q failed", directory)
	}
	return nil
}
