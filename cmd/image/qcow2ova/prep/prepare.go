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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"

	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
)

var hostPartitions = map[string][]string{
	"centos": {"/proc", "/dev", "/sys", "/var/run/", "/etc/machine-id"},
	"rhel":   {"/proc", "/dev", "/sys", "/var/run/", "/etc/machine-id"},
	"fedora": {"/proc", "/dev", "/sys", "/run/", "/etc/machine-id"},
}

// prepare is a function prepares the CentOS or RHEL image for capturing, this includes
// - Installs the cloud-init
// - Install and configure multipath for rootfs
// - Install all the required modules for PowerVM
// - Sets the root password
func prepare(mnt, volume, dist, rhnuser, rhnpasswd, rootpasswd string) error {
	// Setup loop device and cleanup on exit
	lo, err := setupLoop(volume)
	if err != nil {
		return err
	}
	defer removeLoop(lo)

	if err = partprobe(lo); err != nil {
		return err
	}

	partition, err := getPartition(lo)
	if err != nil {
		return err
	}

	partDev := fmt.Sprintf("%sp%s", lo, partition)
	fsType, err := getFSType(partDev)
	if err != nil {
		return err

	}
	switch fsType {
	case "btrfs":
		if err = mount("defaults,subvol=root", partDev, filepath.Join(mnt)); err != nil {
			return err
		}
		defer Umount(mnt)
	case "ext2", "ext3", "ext4", "xfs":
		if err = mount("nouuid", partDev, mnt); err != nil {
			return err
		}
		defer Umount(mnt)
	}

	// Resize partition
	if err = growpart(lo, partition); err != nil {
		return err
	}

	switch fsType {
	case "xfs":
		if err = xfsGrow(partDev); err != nil {
			return err
		}
	case "ext2", "ext3", "ext4":
		if err = resize2fs(partDev); err != nil {
			return err
		}
	case "btrfs":
		if err = growBtrfs(filepath.Join(mnt, "root"), "max"); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unable to handle the %s filesystem for %s", fsType, partDev)
	}

	// Mount boot partition
	fstabPath := filepath.Join(mnt, "etc", "fstab")
	bootMount := filepath.Join(mnt, "boot")
	if deviceuuid, err := bootDeviceuuid(fstabPath); err == nil && deviceuuid != "" {
		if bootDev, err := findDevice(deviceuuid); err == nil {
			if fsType == "btrfs" {
				if err = mount("defaults", bootDev, bootMount); err != nil {
					return err
				}
			} else {
				if err = mount("nouuid", bootDev, bootMount); err != nil {
					return err
				}
			}
			defer Umount(bootMount)
		} else {
			return err
		}
	} else if err != nil {
		return err
	}

	// Verify /boot is mounted properly and files are present.
	bootDirFiles := []string{"config-*.ppc64le", "efi", "grub2", "initramfs-*.ppc64le.img", "loader", "symvers-*.ppc64le.*", "System.map-*.ppc64le", "vmlinuz-*.ppc64le"}
	for _, file := range bootDirFiles {
		exist, err := checkFileExists(filepath.Join(bootMount, file))
		if err != nil {
			return fmt.Errorf("error while validating contents of /boot directory. %v", err)
		}
		if !exist {
			return fmt.Errorf("%s does not exist in the boot directory", file)
		}
	}

	// mount the host partitions
	for _, p := range hostPartitions[dist] {
		err = mount("bind", p, filepath.Join(mnt, p))
		if err != nil {
			return err
		}
	}
	defer UmountHostPartitions(mnt, dist)

	if setupStr, err := Render(dist, rhnuser, rhnpasswd, rootpasswd); err == nil {
		if err = os.WriteFile(filepath.Join(mnt, "setup.sh"), []byte(setupStr), 0744); err != nil {
			return err
		}
	} else {
		return err
	}

	files := map[string]string{
		"/etc/cloud/cloud.cfg":       CloudConfig,
		"/etc/cloud/ds-identify.cfg": dsIdentify,
	}
	for path, content := range files {
		if err := os.WriteFile(filepath.Join(mnt, path), []byte(content), 0644); err != nil {
			return err
		}
	}

	if err = Chroot(mnt); err != nil {
		return err
	}
	defer ExitChroot()

	if err = os.Chdir("/"); err != nil {
		return err
	}

	status, out, errr := utils.RunCMD("/setup.sh")
	if status != 0 {
		return fmt.Errorf("script /setup.sh failed with exitstatus: %d, stdout: %s, stderr: %s", status, out, errr)
	}

	return nil
}

func UmountHostPartitions(mnt, dist string) {
	for _, p := range hostPartitions[dist] {
		Umount(filepath.Join(mnt, p))
	}
}

func Prepare4capture(mnt, volume, dist, rhnuser, rhnpasswd, rootpasswd string) error {
	//cwd, err := os.Getwd()
	//if err != nil {
	//	return err
	//}
	//defer os.Chdir(cwd)
	switch dist := strings.ToLower(dist); dist {
	case "rhel", "centos", "fedora":
		return prepare(mnt, volume, dist, rhnuser, rhnpasswd, rootpasswd)
	case "coreos":
		klog.Info("No image preparation required for the coreos.")
		return nil
	default:
		return fmt.Errorf("not a supported distro: %s", dist)
	}
}
