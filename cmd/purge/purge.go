package purge

import (
	"fmt"
	"github.com/ppc64le-cloud/pvsadm/cmd/purge/images"
	"github.com/ppc64le-cloud/pvsadm/cmd/purge/networks"
	"github.com/ppc64le-cloud/pvsadm/cmd/purge/vms"
	"github.com/ppc64le-cloud/pvsadm/cmd/purge/volumes"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/spf13/cobra"
	"time"
)

var Cmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge the powervs resources",
	Long: `Purge the powervs resources

# Set the API key or feed the --api-key commandline argument
export IBMCLOUD_API_KEY=<IBM_CLOUD_API_KEY>

Examples:
  # Delete all the virtual machines which are created before 4hrs
  pvsadm purge vms --instance-name upstream-core --before 4h

  # Delete all the virtual machines created since 24hrs
  pvsadm purge vms --instance-name upstream-core --since 24h

  # Delete all the volumes which aren't assigned to any virtual machines
  pvsadm purge volumes --instance-name upstream-core

  # Delete all the networks and ignore if any errors during the delete operation
  pvsadm purge networks --instance-name upstream-core --ignore-errors

  # Delete all the images without asking any confirmation
  pvsadm purge images --instance-name upstream-core --no-prompt

  # Delete all the images with debugging logs for IBM cloud APIs
  pvsadm purge images --instance-name upstream-core --debug

  # Delete all the virtual machines starts with k8s-cluster-
  pvsadm purge vms --instance-name upstream-core --regexp "^k8s-cluster-.*"

  # List the purgeable candidate virtual machines and exit without deleting
  pvsadm purge vms --instance-name upstream-core --dry-run
`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if pkg.Options.Since != 0 && pkg.Options.Before != 0 {
			return fmt.Errorf("--since and --before options can not be set at a time")
		}
		if pkg.Options.InstanceID == "" && pkg.Options.InstanceName == "" {
			return fmt.Errorf("--instance-name or --instance-name required")
		}
		return nil
	},
}

func init() {
	Cmd.AddCommand(images.Cmd)
	Cmd.AddCommand(vms.Cmd)
	Cmd.AddCommand(networks.Cmd)
	Cmd.AddCommand(volumes.Cmd)
	Cmd.PersistentFlags().StringVarP(&pkg.Options.InstanceID, "instance-id", "i", "", "Instance ID of the PowerVS instance")
	Cmd.PersistentFlags().StringVarP(&pkg.Options.InstanceName, "instance-name", "n", "", "Instance name of the PowerVS")
	Cmd.PersistentFlags().BoolVar(&pkg.Options.DryRun, "dry-run", false, "dry run the action and don't delete the actual resources")
	Cmd.PersistentFlags().DurationVar(&pkg.Options.Since, "since", 0*time.Second, "Remove resources since mentioned duration(format: 99h99m00s), mutually exclusive with --before")
	Cmd.PersistentFlags().DurationVar(&pkg.Options.Before, "before", 0*time.Second, "Remove resources before mentioned duration(format: 99h99m00s), mutually exclusive with --since")
	Cmd.PersistentFlags().BoolVar(&pkg.Options.NoPrompt, "no-prompt", false, "Show prompt before doing any destructive operations")
	Cmd.PersistentFlags().BoolVar(&pkg.Options.IgnoreErrors, "ignore-errors", false, "Ignore any errors during the operations")
	Cmd.PersistentFlags().StringVar(&pkg.Options.Expr, "regexp", "", "Regular Expressions for filtering the selection")
}
