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

package pkg

import (
	"time"
)

// IsPurgeable is a function decides the candidate is really qualifies to be purged
// Condition:
//   - before and since are both
//   - If no before and since set then return true
//   - If both before and since set then return false
func IsPurgeable(candidate time.Time, before time.Duration, since time.Duration) bool {
	now := time.Now().In(time.UTC)
	start := time.Time{}
	end := now
	if before == 0 && since == 0 {
		return true
	} else if before != 0 && since != 0 {
		return false
	}
	if before != 0 {
		end = now.Add(-before)
	}
	if since != 0 {
		start = now.Add(-since)
	}
	if candidate.After(start) && candidate.Before(end) {
		return true
	}
	return false
}
