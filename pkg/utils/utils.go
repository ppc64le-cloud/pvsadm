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
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

func FormatProcessor(proc *float64) string {
	return strconv.FormatFloat(*proc, 'f', -1, 64)
}

func FormatMemory(memory *float64) string {
	return strconv.FormatFloat(*memory, 'f', -1, 64)
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Ensure that either the workspaceID or the workspaceName is set, along with the API Key.
func EnsurePrerequisitesAreSet(apiKey, workspaceID, workspaceName string) error {
	if apiKey == "" {
		return fmt.Errorf("api-key can't be empty, pass the token via --api-key or set IBMCLOUD_APIKEY environment variable")
	}

	if workspaceID == "" && workspaceName == "" {
		return fmt.Errorf("--workspace-id or --workspace-name required")
	}
	return nil
}

func RetrieveValFromMap[K comparable, V any](m map[K]V, key K) V {
	return m[key]
}

// PollUntil validates if a certain condition is met at defined poll intervals.
// If a timeout is reached, an associated error is returned to the caller.
// condition contains the use-case specific code that returns true when a certain condition is achieved.
func PollUntil(pollInterval, timeOut <-chan time.Time, condition func() (bool, error)) error {
	for {
		select {
		case <-timeOut:
			return fmt.Errorf("timed out while waiting for job to complete")
		case <-pollInterval:
			if done, err := condition(); err != nil {
				return err
			} else if done {
				return nil
			}
		}
	}
}

// SpinnerPollUntil validates if a certain condition is met at defined poll intervals.
// until the condition is met, a loading spinner is displayed
// If a timeout is reached, an associated error is returned to the caller.
// condition contains the use-case specific code that returns true when a certain condition is achieved.
func SpinnerPollUntil(pollInterval, timeOut <-chan time.Time, condition func() (string, bool, error)) error {
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Color("cyan")
	startTime := time.Now()

	var message string
	var done bool
	var err error
	secondTicker := time.NewTicker(1 * time.Second)
	defer secondTicker.Stop()

	_, done, err = condition()
	if err != nil {
		return fmt.Errorf("initial condition check failed: %w", err)
	}
	s.Start()
	for {
		select {
		case <-secondTicker.C:
			s.Suffix = fmt.Sprintf(" %s (Time elapsed: %s)", message, time.Since(startTime).Round(time.Second))
		case <-pollInterval:
			message, done, err = condition()
			if err != nil || done {
				s.Stop()
				return err
			}
		case <-timeOut:
			s.Stop()
			s.Suffix = " timed out while waiting for job to complete"
			return fmt.Errorf("timed out while waiting for job to complete")

		}
	}
}

// Checks if "btrfs" is listed in /proc/filesystems
func IsBtrfsSupported() (bool, error) {
	file, err := os.Open("/proc/filesystems")
	if err != nil {
		return false, fmt.Errorf("failed to open /proc/filesystems: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "btrfs") {
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error reading /proc/filesystems: %v", err)
	}
	return false, nil
}
