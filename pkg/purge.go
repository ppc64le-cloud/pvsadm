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
