package framework

import (
	"fmt"
	"github.com/onsi/ginkgo"
)

// NegativeIt will postfix with the negative tag
func NegativeIt(text string, body interface{}, timeout ...float64) bool {
	return ginkgo.It(text+" [negative]", body, timeout...)
}

// Describe annotates the text with the subcommand label.
func Describe(cmd, text string, body func()) bool {
	return ginkgo.Describe(fmt.Sprintf("[cmd:%s] %s", cmd, text), body)
}
