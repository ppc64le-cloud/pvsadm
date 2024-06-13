// Copyright 2021 IBM Corp
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package purge

import (
	"fmt"
	"time"

	"github.com/ppc64le-cloud/pvsadm/cmd/purge/images"
	"github.com/ppc64le-cloud/pvsadm/cmd/purge/keys"
	"github.com/ppc64le-cloud/pvsadm/cmd/purge/networks"
	"github.com/ppc64le-cloud/pvsadm/cmd/purge/vms"
	"github.com/ppc64le-cloud/pvsadm/cmd/purge/volumes"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge the PowerVS resources",
	Long: `Purge the PowerVS resources

# Set the API key or feed the --api-key commandline argument
export IBMCLOUD_API_KEY=<IBM_CLOUD_API_KEY>

Examples:
  # Delete all the virtual machines which are created before 4hrs
  pvsadm purge vms --workspace-name upstream-core --before 4h

  # Delete all the virtual machines created since 24hrs
  pvsadm purge vms --workspace-name upstream-core --since 24h

  # Delete all the volumes which aren't assigned to any virtual machines
  pvsadm purge volumes --workspace-name upstream-core

  # Delete all the networks and ignore if any errors during the delete operation
  pvsadm purge networks --workspace-name upstream-core --ignore-errors

  # Delete all the networks along with the instances and their assigned ports
  pvsadm purge networks --workspace-name upstream-core --instances true --ports true

  # Delete all the images without asking any confirmation
  pvsadm purge images --workspace-name upstream-core --no-prompt

  # Delete all the images with debugging logs for IBM cloud APIs
  pvsadm purge images --workspace-name upstream-core --debug

  # Delete all the virtual machines starts with k8s-cluster-
  pvsadm purge vms --workspace-name upstream-core --regexp "^k8s-cluster-.*"

  # List the purgeable candidate virtual machines and exit without deleting
  pvsadm purge vms --workspace-name upstream-core --dry-run

  # Delete all the ssh keys which are created before 12hrs
  pvsadm purge keys --workspace-name upstream-core --before 12h --regexp "^rdr-.*"

  # Delete all the ssh keys starts with rdr-
  pvsadm purge keys --workspace-name upstream-core --regexp "^rdr-.*"
`,
	GroupID: "resource",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Code block to execute the strict check mentioned in the rootcmd for the environment.
		// This block is needed as a workaround mentioned in https://github.com/spf13/cobra/issues/252
		// if multiple PersistentPreRunE present in the code
		root := cmd
		for ; root.HasParent(); root = root.Parent() {
		}
		if err := root.PersistentPreRunE(cmd, args); err != nil {
			return err
		}

		if pkg.Options.Since != 0 && pkg.Options.Before != 0 {
			return fmt.Errorf("--since and --before options can not be set at a time")
		}
		if pkg.Options.WorkspaceID == "" && pkg.Options.WorkspaceName == "" {
			return fmt.Errorf("--workspace-id or --workspace-name required")
		}
		return nil
	},
}

func init() {
	Cmd.AddCommand(images.Cmd)
	Cmd.AddCommand(vms.Cmd)
	Cmd.AddCommand(networks.Cmd)
	Cmd.AddCommand(volumes.Cmd)
	Cmd.AddCommand(keys.Cmd)
	Cmd.PersistentFlags().StringVarP(&pkg.Options.WorkspaceID, "instance-id", "i", "", "Instance ID of the PowerVS workspace")
	Cmd.PersistentFlags().MarkDeprecated("instance-id", "instance-id is deprecated, workspace-id should be used")
	Cmd.PersistentFlags().StringVarP(&pkg.Options.WorkspaceName, "instance-name", "n", "", "Instance name of the PowerVS")
	Cmd.PersistentFlags().MarkDeprecated("instance-name", "instance-name is deprecated, workspace-name should be used")
	Cmd.PersistentFlags().StringVarP(&pkg.Options.WorkspaceID, "workspace-id", "", "", "Workspace ID of the PowerVS workspace")
	Cmd.PersistentFlags().StringVarP(&pkg.Options.WorkspaceName, "workspace-name", "", "", "Workspace name of the PowerVS workspace")
	Cmd.PersistentFlags().BoolVar(&pkg.Options.DryRun, "dry-run", false, "dry run the action and don't delete the actual resources")
	Cmd.PersistentFlags().DurationVar(&pkg.Options.Since, "since", 0*time.Second, "Remove resources since mentioned duration(format: 99h99m00s), mutually exclusive with --before")
	Cmd.PersistentFlags().DurationVar(&pkg.Options.Before, "before", 0*time.Second, "Remove resources before mentioned duration(format: 99h99m00s), mutually exclusive with --since")
	Cmd.PersistentFlags().BoolVar(&pkg.Options.NoPrompt, "no-prompt", false, "Show prompt before doing any destructive operations")
	Cmd.PersistentFlags().BoolVar(&pkg.Options.IgnoreErrors, "ignore-errors", false, "Ignore any errors during the operations")
	Cmd.PersistentFlags().StringVar(&pkg.Options.Expr, "regexp", "", "Regular Expressions for filtering the selection")
}
