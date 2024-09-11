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

//go:build !windows

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
