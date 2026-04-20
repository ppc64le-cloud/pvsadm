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
	"fmt"
	"os/exec"
	"strings"

	"k8s.io/klog/v2"
)

var imageBinaries = map[string]string{
	"qemu-img": "yum install qemu-img -y",
	"growpart": "yum install cloud-utils-growpart -y",
}

type Rule struct {
	missingBins []string
}

func (p *Rule) String() string {
	return "tools"
}

func (p *Rule) Verify() error {
	for binary := range imageBinaries {
		path, err := exec.LookPath(binary)
		if err != nil {
			p.missingBins = append(p.missingBins, binary)
		} else {
			klog.Infof("%s found at %s", binary, path)
		}
	}
	if len(p.missingBins) > 0 {
		return fmt.Errorf("executable file(s) not found in $PATH: %s", strings.Join(p.missingBins, ", "))
	}
	return nil
}

func (p *Rule) Hint() string {
	if len(p.missingBins) > 0 {
		var hints []string
		for _, binary := range p.missingBins {
			hints = append(hints, imageBinaries[binary])
		}
		return strings.Join(hints, "\n")
	}
	return ""
}
