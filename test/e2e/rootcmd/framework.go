package rootcmd

import (
	"github.com/ppc64le-cloud/pvsadm/test/e2e/framework"
)

const (
	command = "pvsadm"
)

// CMDDescribe annotates the test with the subcommand label.
func CMDDescribe(text string, body func()) bool {
	return framework.Describe(command, text, body)
}
