package upload

import (
	"fmt"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload the image to the IBM COS",
	Long:  `Upload the image to the IBM COS`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("To Be Implemented...")
		return nil
	},
}
