package validate

import (
	"github.com/ppc64le-cloud/pvsadm/cmd/image/qcow2ova/validate/diskspace"
	"github.com/ppc64le-cloud/pvsadm/cmd/image/qcow2ova/validate/platform"
	"github.com/ppc64le-cloud/pvsadm/cmd/image/qcow2ova/validate/tools"
	"github.com/ppc64le-cloud/pvsadm/cmd/image/qcow2ova/validate/user"
)

func init() {
	//TODO: Add Operating system check
	AddRule(&user.Rule{})
	AddRule(&platform.Rule{})
	AddRule(&tools.Rule{})
	AddRule(&diskspace.Rule{})
}
