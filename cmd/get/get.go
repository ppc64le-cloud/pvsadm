package get

import (
	"github.com/ppc64le-cloud/pvsadm/cmd/get/events"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "get",
	Short: "Get the resources",
	Long:  `Get the resources`,
}

func init() {
	Cmd.AddCommand(events.Cmd)
	Cmd.PersistentFlags().StringVarP(&pkg.Options.InstanceID, "instance-id", "i", "", "Instance ID of the PowerVS instance")
	Cmd.PersistentFlags().StringVarP(&pkg.Options.InstanceName, "instance-name", "n", "", "Instance name of the PowerVS")
}
