package qcow2ova

import (
	"io/ioutil"
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
		image, err = ioutil.TempFile("", "qcow2ova")
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
		Expect(status).To(Equal(0))
		Expect(stderr).To(Equal(""))
		Expect(stdout).To(ContainSubstring("Examples:"))
	})

	framework.NegativeIt("run without image-dist", func() {
		status, _, stderr := runImageQcow2OvaCMD(
			"--image-url", image.Name(),
		)
		Expect(status).NotTo(Equal(0))
		Expect(stderr).To(ContainSubstring("image-dist is a mandatory flag"))
	})
})
