package volumes

import (
	"fmt"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"time"
)

var before time.Duration

const deletePromptMessage = "Deleting all the volumes in available state, volumes can't be claimed back once deleted. Do you really want to continue?"

var Cmd = &cobra.Command{
	Use:   "volumes",
	Short: "Purge the powervs volumes",
	Long:  `Deletes all the volumes for the powervs instance which are in available state(not attached to any instances)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opt := pkg.Options

		c, err := client.NewClient(opt.APIKey)
		if err != nil {
			klog.Error(err)
			return err
		}

		pvmclient, err := client.NewPVMClient(c, opt.InstanceID, opt.InstanceName)
		if err != nil {
			return err
		}
		volumes, err := pvmclient.VolumeClient.GetAllPurgeableByLastUpdateDate(opt.Before, opt.Since)
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
			if opt.NoPrompt || utils.AskYesOrNo(deletePromptMessage) {
				for _, volume := range volumes {
					if *volume.State == "available" {
						klog.Infof("Deleting the %s, and ID: %s", *volume.Name, *volume.VolumeID)
						err = pvmclient.VolumeClient.DeleteVolume(*volume.VolumeID)
						if err != nil && !opt.IgnoreErrors {
							return err
						}
					}
				}
			}
		}
		return nil
	},
}
