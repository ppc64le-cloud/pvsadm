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

var (
	id string
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get DHCP Server",
	Long:  `Get DHCP Server`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opt := pkg.Options

		c, err := client.NewClientWithEnv(opt.APIKey, opt.Environment, opt.Debug)
		if err != nil {
			klog.Errorf("failed to create a session with IBM cloud, err: %v", err)
			return err
		}

		pvmclient, err := client.NewPVMClientWithEnv(c, opt.InstanceID, opt.InstanceName, opt.Environment)
		if err != nil {
			return err
		}

		server, err := pvmclient.DHCPClient.Get(id)
		if err != nil {
			return fmt.Errorf("failed to get a dhcpserver, err: %v", err)
		}
		if server.Network.Name == nil {
			klog.Infof("DHCP client reports status: %s, please retry in sometime.", *server.Status)
			return nil
		}
		var IPandMAC string
		for _, lease := range server.Leases {
			IPandMAC += fmt.Sprintf("%s-%s\n", *lease.InstanceIP, *lease.InstanceMacAddress)
		}
		table := utils.NewTable()
		table.SetHeader([]string{"Network Name", "IP - MAC", "Status"})
		table.Append([]string{*server.Network.Name, IPandMAC, *server.Status})
		table.Table.Render()
		return nil
	},
}

func init() {
	getCmd.Flags().StringVar(&id, "id", "", "Instance ID of the Cloud connection")
	_ = getCmd.MarkFlagRequired("id")
}
