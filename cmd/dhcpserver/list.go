// Copyright 2022 IBM Corp
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

package dhcpserver

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Get PowerVS DHCP servers",
	Long:  `Get PowerVS DHCP servers`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opt := pkg.Options

		c, err := client.NewClientWithEnv(opt.APIKey, opt.Environment, opt.Debug)
		if err != nil {
			klog.Errorf("failed to create a session with IBM cloud: %v", err)
			return err
		}

		pvmclient, err := client.NewPVMClientWithEnv(c, opt.InstanceID, opt.InstanceName, opt.Environment)
		if err != nil {
			return err
		}

		dhcpservers, err := pvmclient.DHCPClient.GetAll()
		if err != nil {
			return fmt.Errorf("failed to get the networks, err: %v", err)
		}

		if len(dhcpservers) == 0 {
			klog.Info("There are no DHCP servers associated with the instance id provided.")
			return nil
		}

		table := utils.NewTable()
		table.SetHeader([]string{"ID", "Network ID", "Network Name", "Status"})
		for _, dhcpserver := range dhcpservers {
			if dhcpserver.Network.ID == nil || dhcpserver.Network.Name == nil {
				// just in case, if the network is not ready, and the DHCP status reports as BUILD.
				// printing the available information must suffice.
				table.Append([]string{*dhcpserver.ID, "", "", *dhcpserver.Status})
				continue
			}
			table.Append([]string{*dhcpserver.ID, *dhcpserver.Network.ID, *dhcpserver.Network.Name, *dhcpserver.Status})
		}
		table.Table.Render()
		return nil

	},
}
