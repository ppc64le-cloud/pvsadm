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
