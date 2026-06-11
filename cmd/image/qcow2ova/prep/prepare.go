// Copyright 2021 IBM Corp
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prep

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"

	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
)

var (
	hostPartitions = []string{"/proc", "/dev", "/sys", "/var/run/", "/etc/machine-id"}
)

// prepare is a function prepares the CentOS or RHEL image for capturing, this includes
// - Installs the cloud-init
// - Install and configure multipath for rootfs
// - Install all the required modules for PowerVM
// - Sets the root password
func prepare(mnt, volume, dist, rhnuser, rhnpasswd, rootpasswd, writeToDirPath string, writeFilesList []string) error {
	lo, err := setupLoop(volume)
	if err != nil {
		return err
	}
	defer removeLoop(lo)

	err = partprobe(lo)
	if err != nil {
		return err
	}

	partition, err := getPartition(lo)
	if err != nil {
		return err
	}

	partDev := lo + "p" + partition

	err = mount("nouuid", partDev, mnt)
	if err != nil {
		return err
	}
	defer Umount(mnt)

	err = growpart(lo, partition)
	if err != nil {
		return err
	}

	fsType, err := getFSType(partDev)
	if err != nil {
		return err
	}

	switch fsType {
	case "xfs":
		err = xfsGrow(partDev)
		if err != nil {
			return err
		}
	case "ext2", "ext3", "ext4":
		err = resize2fs(partDev)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unable to handle the %s filesystem for %s", fsType, partDev)
	}

	fstabPath := filepath.Join(mnt, "etc", "fstab")

	//get the boot partition name
	deviceuuid, err := bootDeviceuuid(fstabPath)
	if err != nil {
		return err
	}

	if deviceuuid != "" {
		bootDev, err := findDevice(deviceuuid)
		if err != nil {
			return err
		}
		err = mount("nouuid", bootDev, filepath.Join(mnt, "boot"))
		if err != nil {
			return err
		}
		defer Umount(filepath.Join(mnt, "boot"))
	}

	// Verify /boot is mounted properly and files are present.
	bootDirFiles := []string{"config-*.ppc64le", "efi", "grub2", "initramfs-*.ppc64le.img", "loader", "symvers-*.ppc64le.*", "System.map-*.ppc64le", "vmlinuz-*.ppc64le"}
	for _, file := range bootDirFiles {
		exist, err := checkFileExists(filepath.Join(mnt, "boot", file))
		if err != nil {
			return fmt.Errorf("error while validating contents of /boot directory. %v", err)
		}
		if !exist {
			return fmt.Errorf("%s does not exist in the boot directory", file)
		}
	}

	// mount the host partitions
	for _, p := range hostPartitions {
		err = mount("bind", p, filepath.Join(mnt, p))
		if err != nil {
			return err
		}
	}
	defer UmountHostPartitions(mnt)

	setupStr, err := Render(dist, rhnuser, rhnpasswd, rootpasswd)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(mnt, "setup.sh"), []byte(setupStr), 0744)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(mnt, "/etc/cloud/cloud.cfg"), []byte(CloudConfig), 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(mnt, "/etc/cloud/ds-identify.cfg"), []byte(dsIdentify), 0644)
	if err != nil {
		return err
	}

	// Write user provided files to given path in the image
	imageWritePath := filepath.Join(mnt, writeToDirPath)
	fileInfo, err := os.Stat(imageWritePath)

	// check if path exists or create it
	if err == nil && !fileInfo.IsDir() {
		return fmt.Errorf("path exists but it is a file: %s", imageWritePath)
	}
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(imageWritePath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", imageWritePath, err)
			}
		} else {
			return err
		}
	}

	for _, filePath := range writeFilesList {
		destinationPath := filepath.Join(imageWritePath, filepath.Base(filePath))
		if err := copyFiles(filePath, destinationPath); err != nil {
			return err
		}
	}

	err = Chroot(mnt)
	if err != nil {
		return err
	}
	defer ExitChroot()

	err = os.Chdir("/")
	if err != nil {
		return err
	}

	status, out, errr := utils.RunCMD("/setup.sh")
	if status != 0 {
		return fmt.Errorf("script /setup.sh failed with exitstatus: %d, stdout: %s, stderr: %s", status, out, errr)
	}

	return nil
}

func UmountHostPartitions(mnt string) {
	for _, p := range hostPartitions {
		Umount(filepath.Join(mnt, p))
	}
}

func Prepare4capture(mnt, volume, dist, rhnuser, rhnpasswd, rootpasswd, writeToDirPath string, writeFilesList []string) error {
	//cwd, err := os.Getwd()
	//if err != nil {
	//	return err
	//}
	//defer os.Chdir(cwd)
	switch dist := strings.ToLower(dist); dist {
	case "rhel", "centos":
		return prepare(mnt, volume, dist, rhnuser, rhnpasswd, rootpasswd, writeToDirPath, writeFilesList)
	case "coreos":
		klog.Info("No image preparation required for the coreos.")
		return nil
	default:
		return fmt.Errorf("not a supported distro: %s", dist)
	}
}

func copyFiles(src, dest string) error {
	fileInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if fileInfo.IsDir() {
		return copyDir(src, dest)
	}

	return copyFile(src, dest)

}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	srcInfo, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}

	return out.Close()
}

func copyDir(src, dest string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dest, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dest, err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, destPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, destPath); err != nil {
				return err
			}
		}
	}
	return nil
}
