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
)

var dhcp string
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete DHCP Server",
	Long:  `Delete DHCP Server`,
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

		err = pvmclient.DHCPClient.Delete(dhcp)
		if err != nil {
			return fmt.Errorf("failed to delete a dhcpserver, err: %v", err)
		}

		klog.Infof("Successfully Deleted a DHCP server")
		return nil
	},
}

func init() {
	deleteCmd.Flags().StringVar(&dhcp, "id", "", "Instance ID of the DHCP server to be deleted")
}
