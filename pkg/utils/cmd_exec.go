// Copyright 2021 IBM Corp
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

const defaultExitCode = 1

func RunCMD(cmd string, args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	c := exec.Command(cmd, args...)

	c.Stdout = io.MultiWriter(os.Stdout, &stdout)
	c.Stderr = io.MultiWriter(os.Stderr, &stderr)
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
