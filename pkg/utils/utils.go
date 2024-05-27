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
	"fmt"
	"strconv"
	"time"
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
