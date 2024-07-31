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

package peravailability

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
)

const powerEdgeRouter = "power-edge-router"

var Cmd = &cobra.Command{
	Use:   "per-availability",
	Short: "List regions that support PER",
	Long:  "List regions that support Power Edge Router (PER)",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if pkg.Options.WorkspaceID == "" && pkg.Options.WorkspaceName == "" {
			return fmt.Errorf("--workspace-id or --workspace-name required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var perEnabledRegions = []string{}
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
		ret, err := pvmclient.DatacenterClient.GetAll()
		if err != nil {
			return err
		}
		var supportsPER bool
		for _, datacenter := range ret.Datacenters {
			if datacenter.Capabilities[powerEdgeRouter] {
				perEnabledRegions = append(perEnabledRegions, *datacenter.Location.Region)
				if pvmclient.Zone == *datacenter.Location.Region {
					supportsPER = true
				}
			}
		}
		if !supportsPER {
			klog.Infof("%s, where the current workspace is present does not support PER.", pvmclient.Zone)
		} else {
			klog.Infof("%s, where the current workspace is present supports PER.", pvmclient.Zone)
		}
		sort.Strings(perEnabledRegions)
		klog.Infof("The following zones/datacenters have support for PER:%v.More information at https://cloud.ibm.com/docs/overview?topic=overview-locations", perEnabledRegions)
		return nil
	},
}
