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

package keys

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
)

var Cmd = &cobra.Command{
	Use:   "keys",
	Short: "Delete PowerVS SSH key(s)",
	Long: `Delete PowerVS SSH key(s) matching regex
pvsadm purge --help for information
`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if pkg.Options.Expr == "" {
			return fmt.Errorf("--regexp is required and shouldn't be empty string")
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		opt := pkg.Options

		if pkg.Options.WorkspaceName != "" {
			klog.Infof("Purge SSH keys for the workspace: %s", pkg.Options.WorkspaceName)
		} else {
			klog.Infof("Purge SSH keys for the workspace ID: %s", pkg.Options.WorkspaceID)
		}

		c, err := client.NewClientWithEnv(opt.APIKey, opt.Environment, opt.Debug)
		if err != nil {
			klog.Errorf("failed to create a session with IBM cloud, err: %v", err)
			return err
		}

		pvmclient, err := client.NewPVMClientWithEnv(c, opt.WorkspaceID, opt.WorkspaceName, opt.Environment)
		if err != nil {
			return err
		}

		keys, err := pvmclient.KeyClient.GetAllPurgeable(pkg.Options.Before, pkg.Options.Since, pkg.Options.Expr)
		if err != nil {
			return fmt.Errorf("failed to get the ssh keys, err: %v", err)
		}

		klog.Infof("keys matched are %s", keys)
		if len(keys) != 0 {
			if opt.NoPrompt || utils.AskConfirmation(fmt.Sprintf(utils.DeletePromptMessage, "keys")) {
				for _, key := range keys {
					err = pvmclient.KeyClient.Delete(key)
					if err != nil {
						if opt.IgnoreErrors {
							klog.Errorf("error occurred while deleting the key: %v", err)
						} else {
							return fmt.Errorf("failed to delete a key, err: %v", err)
						}
					}
					klog.Infof("Successfully deleted a key, id: %s", key)
				}
			}
		}
		return nil
	},
}
