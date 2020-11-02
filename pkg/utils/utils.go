package utils

import "strconv"

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
