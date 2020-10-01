package version

import (
	"fmt"
	"github.com/ppc64le-cloud/pvsadm/pkg/version"
	"runtime"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s, GoVersion: %s\n", version.Get(), runtime.Version())
	},
}
