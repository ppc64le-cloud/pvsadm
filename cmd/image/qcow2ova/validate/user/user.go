package user

import (
	"fmt"
	"os"
)

type Rule struct {
}

func (p *Rule) String() string {
	return "user"
}

func (p *Rule) Verify() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("non-root user is executing the qcow2ova sub-command")
	}
	return nil
}

func (p *Rule) Hint() string {
	return "Expected root user to execute the qcow2ova subcommand"
}
