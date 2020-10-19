package _import

import (
	"fmt"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "import",
	Short: "Import the image into PowerVS instances",
	Long:  `Import the image into PowerVS instances`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("To Be Implemented...")
		return nil
	},
}
