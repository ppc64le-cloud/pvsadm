package validate

import (
	"github.com/ppc64le-cloud/pvsadm/cmd/image/qcow2ova/validate/diskspace"
	image_name "github.com/ppc64le-cloud/pvsadm/cmd/image/qcow2ova/validate/image-name"
	"github.com/ppc64le-cloud/pvsadm/cmd/image/qcow2ova/validate/platform"
	"github.com/ppc64le-cloud/pvsadm/cmd/image/qcow2ova/validate/tools"
	"github.com/ppc64le-cloud/pvsadm/cmd/image/qcow2ova/validate/user"
)

func init() {
	//TODO: Add Operating system check
	AddRule(&platform.Rule{})
	AddRule(&user.Rule{})
	AddRule(&image_name.Rule{})
	AddRule(&tools.Rule{})
	AddRule(&diskspace.Rule{})
}
