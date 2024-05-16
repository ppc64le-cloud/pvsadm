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

	"github.com/IBM/go-sdk-core/v5/core"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
)

var (
	cidr, dns, name, cloudConnectionID string
	snat                               bool
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create DHCP Server",
	Long:  `Create DHCP Server`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opt := pkg.Options

		c, err := client.NewClientWithEnv(opt.APIKey, opt.Environment, opt.Debug)
		if err != nil {
			klog.Errorf("failed to create a session with IBM cloud, err: %v", err)
			return err
		}

		pvmclient, err := client.NewPVMClientWithEnv(c, opt.WorkspaceID, opt.WorkspaceName, opt.Environment)
		if err != nil {
			return err
		}

		body := &models.DHCPServerCreate{
			SnatEnabled: core.BoolPtr(snat),
		}
		if cidr != "" {
			body.Cidr = core.StringPtr(cidr)
		}
		if cloudConnectionID != "" {
			body.CloudConnectionID = core.StringPtr(cloudConnectionID)
		}
		if dns != "" {
			body.DNSServer = core.StringPtr(dns)
		}
		if name != "" {
			body.Name = core.StringPtr(name)
		}
		_, err = pvmclient.DHCPClient.Create(body)
		if err != nil {
			return fmt.Errorf("failed to create a dhcpserver, err: %v", err)
		}

		klog.Info("Successfully created a DHCP server")
		return nil
	},
}

func init() {
	createCmd.Flags().StringVar(&cidr, "cidr", "", "CIDR")
	createCmd.Flags().StringVar(&cloudConnectionID, "cloud-connection-id", "", "Instance ID of the Cloud connection")
	createCmd.Flags().StringVar(&dns, "dns-server", "", "DNS Server")
	createCmd.Flags().StringVar(&name, "name", "", "Name")
	createCmd.Flags().BoolVar(&snat, "snat", true, "SNAT Enabled")
}
