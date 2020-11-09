package validate

import (
	"github.com/ppc64le-cloud/pvsadm/cmd/image/qcow2ova/validate/diskspace"
	"github.com/ppc64le-cloud/pvsadm/cmd/image/qcow2ova/validate/platform"
	"github.com/ppc64le-cloud/pvsadm/cmd/image/qcow2ova/validate/tools"
)

func init() {
	//TODO: Add Operating system check
	AddRule(&platform.Rule{})
	AddRule(&tools.Rule{})
	AddRule(&diskspace.Rule{})
}
