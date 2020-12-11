package vms

import (
	"fmt"
	"github.com/IBM-Cloud/power-go-client/errors"
	"github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM-Cloud/power-go-client/power/client/p_cloud_instances"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/audit"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"strconv"
	"strings"
)

const deletePromptMessage = "Deleting all the above instances, instances can't be claimed back once deleted. Do you really want to continue?"

var Cmd = &cobra.Command{
	Use:   "vms",
	Short: "Purge the powervs vms",
	Long: `Purge the powervs vms!
pvsadm purge --help for information
`,
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

		param := p_cloud_instances.NewPcloudCloudinstancesGetParamsWithTimeout(pkg.TIMEOUT).WithCloudInstanceID(pvmclient.InstanceID)
		resp, err := pvmclient.PISession.Power.PCloudInstances.PcloudCloudinstancesGet(param, ibmpisession.NewAuth(pvmclient.PISession, pvmclient.InstanceID))

		if err != nil || resp.Payload == nil {
			klog.Infof("Failed to perform the operation... %v", err)
			return errors.ToError(err)
		}
		usage := resp.Payload.Usage

		fmt.Println("Usage:")
		tu := utils.NewTable()
		tu.SetHeader([]string{"Instances", "Memory", "Proc Units", "processors", "storage", "storageSSD", "storageStandard"})
		tu.Append([]string{strconv.FormatFloat(*usage.Instances, 'f', -1, 64),
			strconv.FormatFloat(*usage.Memory, 'f', -1, 64),
			strconv.FormatFloat(*usage.ProcUnits, 'f', 1, 64),
			strconv.FormatFloat(*usage.Processors, 'f', -1, 64),
			strconv.FormatFloat(*usage.Storage, 'f', 2, 64),
			strconv.FormatFloat(*usage.StorageSSD, 'f', 2, 64),
			strconv.FormatFloat(*usage.StorageStandard, 'f', 2, 64),
		})
		tu.Table.Render()

		instances, err := pvmclient.InstanceClient.GetAllPurgeable(pkg.Options.Before, pkg.Options.Since, pkg.Options.Expr)
		if err != nil {
			return err
		}

		t := utils.NewTable()
		t.SetHeader([]string{"Name", "IP Addresses", "Image", "CPUS", "RAM", "STATUS", "Creation Date"})
		for _, instance := range instances {
			ins, err := pvmclient.InstanceClient.Get(*instance.PvmInstanceID)
			if err != nil {
				klog.Infof("Error occurred while getting the vm %s", err)
				continue
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
					if err != nil {
						if opt.IgnoreErrors {
							klog.Infof("error occurred while deleting the vm: %v", err)
						} else {
							return err
						}
					}
					audit.Log("vms", "delete", pvmclient.InstanceName+":"+*instance.ServerName)
				}
			}
		}
		return nil
	},
}
