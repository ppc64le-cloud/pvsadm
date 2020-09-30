package vms

import (
	"fmt"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"strings"
)

const deletePromptMessage = "Deleting all the above instances, instances can't be claimed back once deleted. Do you really want to continue?"

var Cmd = &cobra.Command{
	Use:   "vms",
	Short: "Purge the powervs vms",
	Long:  `Purge the powervs vms!`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opt := pkg.Options

		c, err := client.NewClient(opt.APIKey)
		if err != nil {
			return err
		}

		pvmclient, err := client.NewPVMClient(c, opt.InstanceID, opt.InstanceName)
		if err != nil {
			return err
		}

		instances, err := pvmclient.InstanceClient.GetAllPurgeable(pkg.Options.Before, pkg.Options.Since)
		if err != nil {
			return err
		}

		t := utils.NewTable()
		t.SetHeader([]string{"Name", "IP Addresses", "Image", "CPUS", "RAM", "STATUS", "Creation Date"})
		for _, instance := range instances {
			ins, err := pvmclient.InstanceClient.Get(*instance.PvmInstanceID)
			if err != nil {
				return fmt.Errorf("failed to get the instance: %v", err)
			}
			var ipAddrsPrivate, ipAddrsPublic []string
			for _, ip := range ins.Networks {
				if ip.ExternalIP != "" {
					ipAddrsPublic = append(ipAddrsPublic, ip.ExternalIP)
				}
				ipAddrsPrivate = append(ipAddrsPrivate, ip.IPAddress)
			}
			ipString := fmt.Sprintf("External: %s\nPrivate: %s", strings.Join(ipAddrsPublic, ", "), strings.Join(ipAddrsPrivate, ", "))
			status := fmt.Sprintf("Status: %s\nHealth: %s", *instance.Status, instance.Health.Status)
			row := []string{*instance.ServerName, ipString, *instance.ImageID, utils.FormatProcessor(instance.Processors), utils.FormatMemory(instance.Memory), status, instance.CreationDate.String()}
			t.Append(row)
		}
		t.Table.Render()
		if !opt.DryRun && len(instances) != 0 {
			if opt.NoPrompt || utils.AskYesOrNo(deletePromptMessage) {
				for _, instance := range instances {
					klog.Infof("Deleting the %s, and ID: %s", *instance.ServerName, *instance.PvmInstanceID)
					err = pvmclient.InstanceClient.Delete(*instance.PvmInstanceID)
					if err != nil && !opt.IgnoreErrors {
						return err
					}

				}
			}
		}
		return nil
	},
}
