package image

import (
	"fmt"
	_import "github.com/ppc64le-cloud/pvsadm/cmd/image/import"
	"github.com/ppc64le-cloud/pvsadm/cmd/image/qcow2ova"
	"github.com/ppc64le-cloud/pvsadm/cmd/image/upload"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "image",
	Short: "PowerVS Image management",
	Long:  `PowerVS Image management`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("To Be Implemented...")
		return nil
	},
}

func init() {
	Cmd.AddCommand(_import.Cmd)
	Cmd.AddCommand(qcow2ova.Cmd)
	Cmd.AddCommand(upload.Cmd)
}
