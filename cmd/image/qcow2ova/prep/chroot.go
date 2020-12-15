//+build !windows

package prep

import (
	"fmt"
	"os"
	"syscall"
)

var root *os.File

func Chroot(path string) error {
	var err error
	if root != nil {
		return fmt.Errorf("already in chroot")
	}
	root, err = os.Open("/")
	if err != nil {
		return err
	}

	return syscall.Chroot(path)
}

func ExitChroot() error {
	if err := root.Chdir(); err != nil {
		return err
	}
	defer root.Close()

	return syscall.Chroot(".")
}
