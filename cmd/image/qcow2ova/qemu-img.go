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
