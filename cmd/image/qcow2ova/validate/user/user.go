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

package user

import (
	"fmt"
	"os"
)

type Rule struct {
}

func (p *Rule) String() string {
	return "user"
}

func (p *Rule) Verify() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("non-root user is executing the qcow2ova sub-command")
	}
	return nil
}

func (p *Rule) Hint() string {
	return "Expected root user to execute the qcow2ova subcommand"
}
