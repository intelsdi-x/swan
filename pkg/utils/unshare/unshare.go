/*Package unshare when blank imported reexecutes process in isolated pid namespace.*/
package unshare

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

func init() {
	if os.Getenv("UNSHARE_PID_READY") == "" {
		if os.Getuid() != 0 {
			fmt.Println("Error: unshare requires privileged user - please run as root!")
			os.Exit(1)
		}

		runtime.LockOSThread()
		para := []string{"--pid", "--fork", "--mount-proc"}
		args := append(para, os.Args...)

		fp, err := exec.LookPath("unshare")
		if err != nil {
			fmt.Printf("Error: cannot locate unshare binary!")
			os.Exit(1)
		}

		cmd := exec.Command(fp, args...)
		cmd.Env = append(os.Environ(), "UNSHARE_PID_READY=true")
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		err = cmd.Run()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				if waitStatus, ok := exitError.Sys().(syscall.WaitStatus); ok {
					os.Exit(waitStatus.ExitStatus())
				}
			}
			panic(err)
		}
		// No error return from parent process means error code 0.
		os.Exit(0)
	}
}
