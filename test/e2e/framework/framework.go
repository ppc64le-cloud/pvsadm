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

package framework

import (
	"fmt"
	"github.com/onsi/ginkgo"
)

// NegativeIt will postfix with the negative tag
func NegativeIt(text string, body interface{}, timeout ...float64) bool {
	return ginkgo.It(text+" [negative]", body, timeout...)
}

// Describe annotates the text with the subcommand label.
func Describe(cmd, text string, body func()) bool {
	return ginkgo.Describe(fmt.Sprintf("[cmd:%s] %s", cmd, text), body)
}
