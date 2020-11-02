package diskspace

import (
	"fmt"
	"syscall"

	"github.com/ppc64le-cloud/pvsadm/pkg"
	"k8s.io/klog/v2"
)

// Buffer in addition to the mentioned image size
const (
	BUFFER uint64 = 50
	GB     uint64 = 1024 * 1024 * 1024
)

type Rule struct {
}

func (p *Rule) String() string {
	return "diskspace"
}

func (p *Rule) Verify() error {
	opt := pkg.ImageCMDOptions
	var stat syscall.Statfs_t
	err := syscall.Statfs(opt.TempDir, &stat)
	if err != nil {
		return err
	}
	free := (stat.Bavail * uint64(stat.Bsize)) / GB
	need := opt.ImageSize + BUFFER
	klog.Infof("free: %dG, need: %dG", free, need)
	if free < need {
		return fmt.Errorf("%s does not have enough space for the conversion need: %d but got %d", opt.TempDir, need, free)
	}
	return nil
}

func (p *Rule) Hint() string {
	return "make some space in the " + pkg.ImageCMDOptions.TempDir
}
