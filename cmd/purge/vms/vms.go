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

package vms

import (
	"fmt"
	"strings"

	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/audit"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
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

		c, err := client.NewClientWithEnv(opt.APIKey, opt.Environment, opt.Debug)
		if err != nil {
			return err
		}

		pvmclient, err := client.NewPVMClientWithEnv(c, opt.InstanceID, opt.InstanceName, pkg.Options.Environment)
		if err != nil {
			return err
		}

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
