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

	"github.com/spf13/cobra"

	"github.com/ppc64le-cloud/pvsadm/pkg"
)

var Cmd = &cobra.Command{
	Use:   "dhcpserver",
	Short: "dhcpserver command",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if pkg.Options.APIKey == "" {
			return fmt.Errorf("api-key can't be empty, pass the token via --api-key or set IBMCLOUD_API_KEY environment variable")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(deleteCmd)

	Cmd.PersistentFlags().StringVarP(&pkg.Options.InstanceID, "instance-id", "i", "", "Instance ID of the PowerVS instance")
	_ = Cmd.MarkPersistentFlagRequired("instance-id")
}
