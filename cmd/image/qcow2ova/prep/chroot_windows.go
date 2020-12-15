package prep

import (
	"fmt"
)

func Chroot(path string) error {
	return fmt.Errorf("Not supported on windows platform")
}

func ExitChroot() error {
	return fmt.Errorf("Not supported on windows platform")
}
