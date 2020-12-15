package diskspace

import (
	"fmt"
)

type Rule struct {
}

func (p *Rule) String() string {
	return "diskspace"
}

func (p *Rule) Verify() error {
	return fmt.Errorf("Not supported on Windows platform")
}

func (p *Rule) Hint() string {
	return "Please retry on linux/ppc64le platform"
}
