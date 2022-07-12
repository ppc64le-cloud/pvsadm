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
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"
)

var (
	hostPartitions = []string{"/proc", "/dev", "/sys", "/var/run/", "/etc/machine-id"}
)

//prepare is a function prepares the CentOS or RHEL image for capturing, this includes
// - Installs the cloud-init
// - Install and configure multipath for rootfs
// - Install all the required modules for PowerVM
// - Sets the root password
func prepare(mnt, volume, dist, rhnuser, rhnpasswd, rootpasswd string) error {
	lo, err := setupLoop(volume)
	if err != nil {
		return err
	}

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
	err = ioutil.WriteFile(filepath.Join(mnt, "setup.sh"), []byte(setupStr), 0744)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(mnt, "/etc/cloud/cloud.cfg"), []byte(cloudConfig), 0644)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(mnt, "/etc/cloud/ds-identify.cfg"), []byte(dsIdentify), 0644)
	if err != nil {
		return err
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

func Prepare4capture(mnt, volume, dist, rhnuser, rhnpasswd, rootpasswd string) error {
	//cwd, err := os.Getwd()
	//if err != nil {
	//	return err
	//}
	//defer os.Chdir(cwd)
	switch dist := strings.ToLower(dist); dist {
	case "rhel", "centos":
		return prepare(mnt, volume, dist, rhnuser, rhnpasswd, rootpasswd)
	case "coreos":
		klog.Infof("No image preparation required for the coreos...")
		return nil
	default:
		return fmt.Errorf("not a supported distro: %s", dist)
	}
}
