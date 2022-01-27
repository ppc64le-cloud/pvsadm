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

const deletePromptMessage = "Deleting all the above ssh key/key's, Do you really want to continue?"

var Cmd = &cobra.Command{
	Use:   "keys",
	Short: "Delete PowerVS ssh key/keys",
	Long: `Delete PowerVS ssh key/keys matching regex
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

		c, err := client.NewClientWithEnv(opt.APIKey, opt.Environment, opt.Debug)
		if err != nil {
			klog.Errorf("failed to create a session with IBM cloud: %v", err)
			return err
		}

		pvmclient, err := client.NewPVMClientWithEnv(c, opt.InstanceID, opt.InstanceName, opt.Environment)
		if err != nil {
			return err
		}

		keys, err := pvmclient.KeyClient.GetAllPurgeable(pkg.Options.Before, pkg.Options.Since, pkg.Options.Expr)
		if err != nil {
			return fmt.Errorf("failed to get the ssh keys, err: %v", err)
		}

		klog.Infof("keys matched are %s", keys)
		if len(keys) != 0 {
			if opt.NoPrompt || utils.AskConfirmation(deletePromptMessage) {
				for _, key := range keys {
					err = pvmclient.KeyClient.Delete(key)
					if err != nil {
						if opt.IgnoreErrors {
							klog.Infof("error occurred while deleting the key: %v", err)
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
