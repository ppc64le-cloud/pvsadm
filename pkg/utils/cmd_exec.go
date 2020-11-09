package utils

import (
	"bytes"
	"os/exec"
)

const defaultExitCode = 1

func RunCMD(cmd string, args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	c := exec.Command(cmd, args...)

	c.Stdout = &stdout
	c.Stderr = &stderr
	if err := c.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode(), stdout.String(), stderr.String()
		} else {
			// This case is for macOS, exit code could get and stderr will be empty string
			errString := stderr.String()
			if errString == "" {
				errString = err.Error()
			}
			return defaultExitCode, stdout.String(), errString
		}
	}
	return 0, stdout.String(), stderr.String()
}
