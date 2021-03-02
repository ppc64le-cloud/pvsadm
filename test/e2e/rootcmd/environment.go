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
		Expect(status).NotTo(Equal(0))
		Expect(stderr).To(ContainSubstring("invalid \"fake\" IBM Cloud Environment passed"))
	})
})
