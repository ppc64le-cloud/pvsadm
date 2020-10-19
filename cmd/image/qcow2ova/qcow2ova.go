package qcow2ova

import (
	"fmt"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "qcow2ova",
	Short: "Convert the qcow2 image to ova format",
	Long:  `Convert the qcow2 image to ova format`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("To Be Implemented...")
		return nil
	},
}
