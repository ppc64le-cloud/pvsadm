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

	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/audit"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

const deletePromptMessage = "Deleting all the above networks, networks can't be claimed back once deleted. Do you really want to continue?"

var (
	deletePorts, deleteInstances bool
)

var Cmd = &cobra.Command{
	Use:   "networks",
	Short: "Purge the powervs networks",
	Long: `Purge the powervs networks!
pvsadm purge --help for information
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opt := pkg.Options

		c, err := client.NewClientWithEnv(opt.APIKey, opt.Environment, opt.Debug)
		if err != nil {
			return err
		}

		pvmclient, err := client.NewPVMClientWithEnv(c, opt.InstanceID, opt.InstanceName, pkg.Options.Environment)
		if err != nil {
			return err
		}
		klog.Infof("Purging the networks for the instance: %v", pvmclient.InstanceID)

		networks, err := pvmclient.NetworkClient.GetAllPurgeable(opt.Before, opt.Since, opt.Expr)
		if err != nil {
			return fmt.Errorf("failed to get the list of networks: %v", err)
		}
		table := utils.NewTable()

		table.Render(networks, []string{"href"})
		if !opt.DryRun && len(networks) != 0 {
			if opt.NoPrompt || utils.AskYesOrNo(deletePromptMessage) {
				for _, network := range networks {
					if deleteInstances || deletePorts {
						ports, err := pvmclient.NetworkClient.GetAllPorts(*network.NetworkID)
						if err != nil {
							return fmt.Errorf("failed to get the list of ports: %v", err)
						}

						// Clean up instances and ports associated with the network instance
						for _, port := range ports.Ports {
							pvminstance := port.PvmInstance
							portID := port.PortID
							if deleteInstances && pvminstance != nil {
								err = pvmclient.InstanceClient.Delete(pvminstance.PvmInstanceID)
								if err != nil {
									return fmt.Errorf("failed to delete the instance: %v", err)
								}
								klog.Infof("Successfully deleted a instance %s using network '%s'", pvminstance.PvmInstanceID, *network.Name)
							}
							if deletePorts {
								err = pvmclient.NetworkClient.DeletePort(*network.NetworkID, *portID)
								if err != nil {
									return fmt.Errorf("failed to delete a port, err: %v", err)
								}
								klog.Infof("Successfully deleted a port %s using network '%s'", *portID, *network.Name)
							}
						}
					}
					klog.Infof("Deleting the %s, and ID: %s", *network.Name, *network.NetworkID)
					err = pvmclient.NetworkClient.Delete(*network.NetworkID)
					if err != nil {
						if opt.IgnoreErrors {
							klog.Infof("error occurred while deleting the network: %v", err)
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
	Cmd.PersistentFlags().BoolVar(&deletePorts, "ports", false, "Delete ports")
	Cmd.PersistentFlags().BoolVar(&deleteInstances, "instances", false, "Delete instances")
}
