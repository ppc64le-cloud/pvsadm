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

package platform

import (
	"fmt"
	"runtime"
)

type Rule struct {
}

func (p *Rule) String() string {
	return "platform"
}

func (p *Rule) Verify() error {
	if runtime.GOOS != "linux" && runtime.GOARCH != "ppc64le" {
		return fmt.Errorf("unsupported os: %s, platform: %s", runtime.GOOS, runtime.GOARCH)
	}
	return nil
}

func (p *Rule) Hint() string {
	return "supported only on linux/ppc64le platform, please run it on RHEL/CentOS(ppc64le)"
}
