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

package tools

import (
	"os/exec"
	"strings"

	"github.com/ppc64le-cloud/pvsadm/pkg"
	"k8s.io/klog/v2"
)

var commands = map[string]string{
	"qemu-img": "yum install qemu-img -y",
	"growpart": "yum install cloud-utils-growpart -y",
}

type Rule struct {
	failedCommand string
}

func (p *Rule) String() string {
	return "tools"
}

func (p *Rule) Verify() error {
	if strings.ToLower(pkg.ImageCMDOptions.ImageDist) == "fedora" {
		commands["btrfs"] = "yum install btrfs-progs -y"
	}
	for command := range commands {
		path, err := exec.LookPath(command)
		if err != nil {
			p.failedCommand = command
			return err
		}
		klog.Infof("%s found at %s", command, path)
	}
	return nil
}

func (p *Rule) Hint() string {
	if p.failedCommand != "" {
		return commands[p.failedCommand]
	}
	return ""
}
