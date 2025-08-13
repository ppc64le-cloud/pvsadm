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
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"

	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
)

const losetupCMD = "losetup"

// setupLoop allocates the free loop device for the backing store
func setupLoop(file string) (string, error) {
	exitcode, out, err := utils.RunCMD(losetupCMD, "-f", "--show", file)
	if exitcode != 0 {
		return "", fmt.Errorf("failed to setup a loop device for file: %s, exitcode: %d, stdout: %s, err: %s", file, exitcode, out, err)
	}

	return strings.TrimSpace(out), nil
}

func removeLoop(loopPath string) error {
	exitcode, stdout, stderr := utils.RunCMD(losetupCMD, "-d", loopPath)
	if exitcode != 0 {
		return fmt.Errorf("failed to remove loop device: %s, exitcode: %d, stdout: %s, err: %s", loopPath, exitcode, stdout, stderr)
	}
	return nil
}

func partprobe(device string) error {
	exitcode, out, err := utils.RunCMD("partprobe", device)
	if exitcode != 0 {
		return fmt.Errorf("failed to partprobe: %s, exitcode: %d, stdout: %s, err: %s", device, exitcode, out, err)
	}
	return nil
}

// get partition number of the image
func getPartition(device string) (string, error) {
	args := fmt.Sprintf("fdisk -l %s | grep ^/dev | wc -l", device)
	exitcode, out, err := utils.RunCMD("bash", "-c", args)
	if exitcode != 0 {
		return "", fmt.Errorf("failed to get partition for device: %s, exitcode: %d, stdout: %s, err: %s", device, exitcode, out, err)
	}
	return strings.TrimSpace(out), nil
}

// growpart resizes the partition
func growpart(device, partition string) error {
	exitcode, out, err := utils.RunCMD("growpart", device, partition)
	if exitcode != 0 {
		return fmt.Errorf("failed to growpart for the device: %s and partition: %s, exitcode: %d, stdout: %s, err: %s", device, partition, exitcode, out, err)
	}
	return nil
}

// getFSType returns the filesystem type for the given device and the partition
func getFSType(device string) (string, error) {
	exitcode, out, err := utils.RunCMD("blkid", device, "-o", "value", "-s", "TYPE")
	if exitcode != 0 {
		return "", fmt.Errorf("failed to get the filesystem type for the device: %s, exitcode: %d, stdout: %s, err: %s", device, exitcode, out, err)
	}
	return strings.TrimSpace(out), nil
}

func xfsGrow(device string) error {
	exitcode, out, err := utils.RunCMD("xfs_growfs", "-d", device)
	if exitcode != 0 {
		return fmt.Errorf("failed to xfs_growfs device: %s, exitcode: %d, stdout: %s, err: %s", device, exitcode, out, err)
	}
	return nil
}

func resize2fs(device string) error {
	exitcode, out, err := utils.RunCMD("resize2fs", device)
	if exitcode != 0 {
		return fmt.Errorf("failed to resize2fs the device: %s, exitcode: %d, stdout: %s, err: %s", device, exitcode, out, err)
	}
	return nil
}

// growBtrfs resizes the mounted Btrfs volume to max size
func growBtrfs(mountPoint string, size string) error {
	exitcode, out, err := utils.RunCMD("btrfs", "filesystem", "resize", size, mountPoint)
	if exitcode != 0 {
		return fmt.Errorf("failed to grow btrfs volume: %s, exitcode: %d, stdout: %s, err: %s", mountPoint, exitcode, out, err)
	}
	return nil
}

func mount(opts, src, target string) error {
	exitcode, out, err := utils.RunCMD("mount", "-o", opts, src, target)
	if exitcode != 0 {
		return fmt.Errorf("failed to bind mount, source: %s and target: %s, exitcode: %d, stdout: %s, err: %s", src, target, exitcode, out, err)
	}
	return nil
}

// checkFileExists return True if file with the name exists else False
func checkFileExists(name string) (bool, error) {
	matches, err := filepath.Glob(name)
	return len(matches) > 0, err
}

// bootDeviceuuid takes fstab file path as input and returns uuid of boot device
func bootDeviceuuid(file string) (string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	deviceuuid := ""
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "boot") {
			fields := strings.Fields(line)
			deviceuuid = strings.Split(fields[0], "=")[1]
		}
	}
	return deviceuuid, nil
}

// findDevice takes UUID as input and returns the device/partition associated with it.
func findDevice(uuid string) (string, error) {
	exitcode, out, err := utils.RunCMD("blkid", "--uuid", uuid)
	if exitcode != 0 {
		return "", fmt.Errorf("failed to get the device name, exitcode: %d, stdout: %s, err: %s", exitcode, out, err)
	}
	return strings.TrimSpace(out), nil
}

func Umount(dir string) error {
	exitcode, out, err := utils.RunCMD("umount", dir)
	if exitcode != 0 {
		if strings.Contains(err, "not mounted") {
			klog.V(1).Infof("Ignoring 'not mounted' error for %s", dir)
			return nil
		}
		if strings.Contains(err, "no mount point specified") {
			klog.V(1).Infof("Ignoring 'no mount point specified' error for %s", dir)
			return nil
		}
		if strings.Contains(err, "target is busy") {
			for retry := 0; retry < 5; retry++ {
				exitcode, _, err = utils.RunCMD("umount", dir)
				if exitcode == 0 || strings.Contains(err, "no mount point specified") || strings.Contains(err, "not mounted") {
					return nil
				}
			}
			klog.V(1).Infof("As '%s' is busy, unmounting it using lazy unmount", dir)
			exitcode, out, err = utils.RunCMD("umount", "-lf", dir)
			if exitcode == 0 {
				return nil
			}
		}
		return fmt.Errorf("failed to unmount, dir: %s, exitcode: %d, stdout: %s, err: %s", dir, exitcode, out, err)
	}
	return nil
}
