package image_name

// Validate if same image name exist in the folder, don't overwrite
import (
	"fmt"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"os"
)

type Rule struct {
}

func (p *Rule) String() string {
	return "image-name"
}

func (p *Rule) Verify() error {
	if _, err := os.Stat(pkg.ImageCMDOptions.ImageName + ".ova.gz"); !os.IsNotExist(err) {
		return fmt.Errorf("file already exist with name: %s", pkg.ImageCMDOptions.ImageName+".ova.gz")
	}
	return nil
}

func (p *Rule) Hint() string {
	return "Please choose a different image-name and retry"
}
