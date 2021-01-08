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

package diskspace

import (
	"fmt"
)

type Rule struct {
}

func (p *Rule) String() string {
	return "diskspace"
}

func (p *Rule) Verify() error {
	return fmt.Errorf("Not supported on Windows platform")
}

func (p *Rule) Hint() string {
	return "Please retry on linux/ppc64le platform"
}
