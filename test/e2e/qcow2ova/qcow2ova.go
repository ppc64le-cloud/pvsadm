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
	"os"

	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"github.com/ppc64le-cloud/pvsadm/test/e2e/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func runImageQcow2OvaCMD(args ...string) (int, string, string) {
	args = append([]string{"image", "qcow2ova"}, args...)
	return utils.RunCMD("pvsadm", args...)
}

var _ = CMDDescribe("pvsadm qcow2ova tests", func() {
	var (
		image *os.File
	)

	BeforeSuite(func() {
		var err error
		Expect(os.Getenv("IBMCLOUD_API_KEY")).NotTo(BeEmpty(), "IBMCLOUD_API_KEY must be set before running \"make test\"")
		image, err = os.CreateTemp("", "qcow2ova")
		Expect(err).NotTo(HaveOccurred())
		_, err = image.WriteString("some dummy image")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterSuite(func() {
		image.Close()
	})

	It("run with --help option", func() {
		status, stdout, stderr := runImageQcow2OvaCMD(
			"--help",
		)
		Expect(status).To(BeZero())
		Expect(stderr).To(BeEmpty())
		Expect(stdout).To(ContainSubstring("Examples:"))
	})

	framework.NegativeIt("run without image-dist", func() {
		status, _, stderr := runImageQcow2OvaCMD(
			"--image-url", image.Name(),
		)
		Expect(status).NotTo(BeZero())
		Expect(stderr).To(ContainSubstring("image-dist is a mandatory flag"))
	})
})
