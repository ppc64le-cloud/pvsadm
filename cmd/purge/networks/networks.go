package networks

import (
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var Cmd = &cobra.Command{
	Use:   "networks",
	Short: "Purge the powervs networks",
	Long:  `Purge the powervs networks!`,
	Run: func(cmd *cobra.Command, args []string) {
		klog.Info("Yet to come...")
	},
}
