// Copyright 2024 IBM Corp
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

package cloudconnections

import (
	"strings"

	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
)

// Struct to contain the information related to a cloud connection.
type cloudConnectionDetails struct {
	name, cloudConnId, state, zone string
	workspaces                     map[string]struct{}
}

// Struct to contain the information related to a workspace.
type workspaceDetails struct {
	name, guid string
}

// Map to contain the zone and their associated workspaces it contains at an account level.
var zoneWorkspaces = make(map[string][]workspaceDetails)

var Cmd = &cobra.Command{
	Use:   "cloud-connections",
	Short: "List the existing cloud connections in the account",
	Long:  "List the existing cloud connections enabled across all workspaces in the account",
	RunE: func(cmd *cobra.Command, args []string) error {
		opt := pkg.Options
		c, err := client.NewClientWithEnv(opt.APIKey, opt.Environment, opt.Debug)
		if err != nil {
			klog.Errorf("failed to create a session with IBM cloud: %v", err)
			return err
		}

		environment, err := client.GetEnvironment(opt.Environment)
		if err != nil {
			return err
		}
		// Retrieve all workspaces that are available in the account.
		workspaceInstances, err := c.ListWorkspaceInstances()
		if err != nil {
			return err
		}
		for _, workspaceInstance := range workspaceInstances.Resources {
			if _, exists := zoneWorkspaces[*workspaceInstance.RegionID]; !exists {
				zoneWorkspaces[*workspaceInstance.RegionID] = make([]workspaceDetails, 0)
			}
			zoneWorkspaces[*workspaceInstance.RegionID] = append(zoneWorkspaces[*workspaceInstance.RegionID], workspaceDetails{name: *workspaceInstance.Name, guid: *workspaceInstance.GUID})
		}
		cloudConnections := map[string]cloudConnectionDetails{}
		authenticator := &core.IamAuthenticator{ApiKey: pkg.Options.APIKey, URL: environment[client.TPEndpoint]}
		klog.Info("Listing cloud connections across all workspaces, please wait..")
		// Create a IBM PI Session per zone and reuse them across the workspaces in the same zone.
		for workspaceZone, workspaces := range zoneWorkspaces {
			pvmclientOptions := ibmpisession.IBMPIOptions{
				Authenticator: authenticator,
				Debug:         pkg.Options.Debug,
				URL:           environment[client.PIEndpoint],
				UserAccount:   c.User.Account,
				Zone:          workspaceZone,
			}
			piSession, err := ibmpisession.NewIBMPISession(&pvmclientOptions)
			if err != nil {
				return err
			}
			// Iterate over the workspaces available in the zone.
			for _, workspace := range workspaces {
				pvmClient, err := client.NewGenericPVMClient(c, workspace.guid, piSession)
				if err != nil {
					return err
				}
				// Retrieve all the cloud connections that are associated with the workspaces.
				cloudConnectionResp, err := pvmClient.CloudConnectionClient.GetAll()
				if err != nil {
					return err
				}
				for _, cloudConnection := range cloudConnectionResp.CloudConnections {
					// a single CC may used across multiple workspaces, the workspaces need to be grouped accordingly.
					if _, exists := cloudConnections[*cloudConnection.Name]; !exists {
						// only add unique CCs and their associated workspaces.
						ccInstance := cloudConnectionDetails{
							name:        *cloudConnection.Name,
							cloudConnId: *cloudConnection.CloudConnectionID,
							state:       *cloudConnection.LinkStatus,
							workspaces:  make(map[string]struct{}),
							zone:        workspaceZone,
						}
						ccInstance.workspaces[workspace.name] = struct{}{}
						cloudConnections[*cloudConnection.Name] = ccInstance
						continue
					}
					// update the map to reuse details of CC, which are used by other workspaces.
					cloudConnections[*cloudConnection.Name].workspaces[workspace.name] = struct{}{}
				}
			}
		}
		if len(cloudConnections) > 0 {
			table := utils.NewTable()
			table.SetHeader([]string{"Cloud Connection ID", "Name", "State", "Workspaces", "Zone"})
			for _, cloudConnection := range cloudConnections {
				workspaces := make([]string, 0, len(cloudConnection.workspaces))
				for workspace := range cloudConnection.workspaces {
					workspaces = append(workspaces, workspace)
				}
				services := strings.Join(workspaces, ",")
				table.Append([]string{cloudConnection.cloudConnId, cloudConnection.name, cloudConnection.state, services, cloudConnection.zone})
			}
			table.Table.Render()
			return nil
		}
		klog.Info("There are no active cloud connections in this account.")
		return nil
	},
}
