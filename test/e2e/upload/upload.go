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

package upload

import (
	"io/ioutil"
	"os"

	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"github.com/ppc64le-cloud/pvsadm/test/e2e/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func runImageUploadCMD(args ...string) (int, string, string) {
	args = append([]string{"image", "upload"}, args...)
	return utils.RunCMD("pvsadm", args...)
}

var _ = CMDDescribe("pvsadm upload tests", func() {
	var (
		image *os.File
		err   error
	)

	image, err = ioutil.TempFile("", "upload")
	Expect(err).NotTo(HaveOccurred())
	_, err = image.WriteString("some dummy image file")
	Expect(err).NotTo(HaveOccurred())
	defer image.Close()

	It("run with --help option", func() {
		status, stdout, stderr := runImageUploadCMD(
			"--help",
		)
		Expect(status).To(Equal(0))
		Expect(stderr).To(Equal(""))
		Expect(stdout).To(ContainSubstring("Examples:"))
	})

	framework.NegativeIt("run without bucket", func() {
		status, _, stderr := runImageUploadCMD(
			"--file", image.Name(),
		)
		Expect(status).NotTo(Equal(0))
		Expect(stderr).To(ContainSubstring("required flag(s) \"bucket\" not set"))
	})
})
