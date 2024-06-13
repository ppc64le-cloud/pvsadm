// Copyright 2022 IBM Corp
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

package rootcmd

import (
	. "github.com/onsi/gomega"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"github.com/ppc64le-cloud/pvsadm/test/e2e/framework"
)

func runCMD(args ...string) (int, string, string) {
	return utils.RunCMD("pvsadm", args...)
}

var _ = CMDDescribe("run pvsadm with", func() {

	framework.NegativeIt("wrong environment", func() {
		status, _, stderr := runCMD(
			"purge",
			"images",
			"-n", "fake_powervs_instance",
			"--env", "fake",
		)
		Expect(status).NotTo(BeZero())
		Expect(stderr).To(ContainSubstring("invalid \"fake\" IBM Cloud Environment passed"))
	})
})
