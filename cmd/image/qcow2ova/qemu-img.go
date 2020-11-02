package qcow2ova

import (
	"fmt"

	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
)

const QemuCMD = "qemu-img"

// qemuImgConvertQcow2Raw converts qcow2 format to RAW
func qemuImgConvertQcow2Raw(source, target string) error {
	args := []string{"convert", "-f", "qcow2", "-O", "raw", source, target}
	exit, out, err := utils.RunCMD(QemuCMD, args...)
	if exit != 0 {
		return fmt.Errorf("failed to convert Qcow2(%s) image to RAW(%s) format, exited with: %d, out: %s, err: %s", source, target, exit, out, err)
	}
	return nil
}

// qemuImgResize resizes the image
func qemuImgResize(image string, size string) error {
	args := []string{"resize", image, size}
	exit, out, err := utils.RunCMD(QemuCMD, args...)
	if exit != 0 {
		return fmt.Errorf("failed to resize image(%s), exited with: %d, out: %s, err: %s", image, exit, out, err)
	}
	return nil
}
