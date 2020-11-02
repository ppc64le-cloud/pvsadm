package platform

import (
	"fmt"
	"runtime"
)

type Rule struct {
}

func (p *Rule) String() string {
	return "platform"
}

func (p *Rule) Verify() error {
	if runtime.GOOS != "linux" && runtime.GOARCH != "ppc64le" {
		return fmt.Errorf("unsupported os: %s, platform: %s", runtime.GOOS, runtime.GOARCH)
	}
	return nil
}

func (p *Rule) Hint() string {
	return "supported only on linux/ppc64le platform"
}
