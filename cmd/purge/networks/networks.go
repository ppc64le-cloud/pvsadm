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

package networks

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/audit"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
)

var (
	deletePorts, deleteInstances bool
)

var Cmd = &cobra.Command{
	Use:   "networks",
	Short: "Purge the PowerVS networks",
	Long: `Purge the PowerVS networks!
pvsadm purge --help for information
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opt := pkg.Options

		c, err := client.NewClientWithEnv(opt.APIKey, opt.Environment, opt.Debug)
		if err != nil {
			return err
		}

		pvmclient, err := client.NewPVMClientWithEnv(c, opt.WorkspaceID, opt.WorkspaceName, pkg.Options.Environment)
		if err != nil {
			return err
		}
		if pkg.Options.WorkspaceName != "" {
			klog.Infof("Purge networks for the workspace: %s", pkg.Options.WorkspaceName)
		} else {
			klog.Infof("Purge networks for the workspace ID: %s", pkg.Options.WorkspaceID)
		}

		networks, err := pvmclient.NetworkClient.GetAllPurgeable(opt.Expr)
		if err != nil {
			return fmt.Errorf("failed to get the list of networks: %v", err)
		}
		table := utils.NewTable()

		table.Render(networks, []string{"href"})
		if !opt.DryRun && len(networks) != 0 {
			if opt.NoPrompt || utils.AskConfirmation(fmt.Sprintf(utils.DeletePromptMessage, "networks")) {
				for _, network := range networks {
					if deleteInstances || deletePorts {
						ports, err := pvmclient.NetworkClient.GetAllPorts(*network.NetworkID)
						if err != nil {
							return fmt.Errorf("failed to get the list of ports: %v", err)
						}

						// Clean up instances and ports associated with the network instance
						for _, port := range ports.Ports {
							pvminstance := port.PvmInstance
							if deleteInstances && (pvminstance != nil) {
								err = pvmclient.InstanceClient.Delete(pvminstance.PvmInstanceID)
								if err != nil {
									if opt.IgnoreErrors {
										klog.Errorf("error occurred while deleting PVMInstance: %s associated with network %s : %v", pvminstance.PvmInstanceID, *network.Name, err)
									} else {
										return err
									}
								}
								klog.Infof("Successfully deleted a instance %s using network '%s'", pvminstance.PvmInstanceID, *network.Name)
							}
							if deletePorts {
								err = pvmclient.NetworkClient.DeletePort(*network.NetworkID, *port.PortID)
								if err != nil {
									if opt.IgnoreErrors {
										klog.Errorf("error occurred while deleting port: %s associated with network %s : %v", *port.PortID, *network.Name, err)
									} else {
										return err
									}
								}
								klog.Infof("Successfully deleted a port %s using network '%s'", *port.PortID, *network.Name)
							}
						}
					}
					klog.Infof("Deleting network: %s with ID: %s", *network.Name, *network.NetworkID)
					err = pvmclient.NetworkClient.Delete(*network.NetworkID)
					if err != nil {
						if opt.IgnoreErrors {
							klog.Errorf("error occurred while deleting the network: %v", err)
						} else {
							return err
						}
					}
					audit.Log("networks", "delete", pvmclient.InstanceName+":"+*network.Name)
				}
			}
		}
		return nil
	},
}

func init() {
	Cmd.PersistentFlags().BoolVar(&deletePorts, "ports", false, "Delete ports that are associated with the network")
	Cmd.PersistentFlags().BoolVar(&deleteInstances, "instances", false, "Delete instances that are associated with the network")
}
