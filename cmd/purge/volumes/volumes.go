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

package volumes

import (
	"fmt"
	"time"

	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/audit"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var before time.Duration

const deletePromptMessage = "Deleting all the volumes in available state, volumes can't be claimed back once deleted. Do you really want to continue?"

var Cmd = &cobra.Command{
	Use:   "volumes",
	Short: "Purge the PowerVS volumes",
	Long: `Deletes all the volumes for the PowerVS instance which are in available state(not attached to any instances)
pvsadm purge --help for information
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opt := pkg.Options

		c, err := client.NewClientWithEnv(opt.APIKey, opt.Environment, opt.Debug)
		if err != nil {
			klog.Error(err)
			return err
		}

		pvmclient, err := client.NewPVMClientWithEnv(c, opt.WorkspaceID, opt.WorkspaceName, pkg.Options.Environment)
		if err != nil {
			return err
		}
		volumes, err := pvmclient.VolumeClient.GetAllPurgeableByLastUpdateDate(opt.Before, opt.Since, opt.Expr)
		if err != nil {
			return fmt.Errorf("failed to get the list of volumes: %v", err)
		}

		t := utils.NewTable()
		t.SetHeader([]string{"Name", "Volume ID", "State", "Last Update Date"})
		for _, volume := range volumes {
			t.Append([]string{*volume.Name, *volume.VolumeID, *volume.State, volume.LastUpdateDate.String()})
		}
		t.Table.Render()

		if !opt.DryRun && len(volumes) != 0 {
			klog.Infof("Deleting all the volumes in available state")
			if opt.NoPrompt || utils.AskConfirmation(deletePromptMessage) {
				for _, volume := range volumes {
					if *volume.State == "available" {
						klog.Infof("Deleting volume: %s with ID: %s", *volume.Name, *volume.VolumeID)
						err = pvmclient.VolumeClient.DeleteVolume(*volume.VolumeID)
						if err != nil {
							if opt.IgnoreErrors {
								klog.Errorf("error occurred while deleting the volume: %v", err)
							} else {
								return err
							}
						}
						audit.Log("volumes", "delete", pvmclient.InstanceName+":"+*volume.Name)
					}
				}
			}
		}
		return nil
	},
}
