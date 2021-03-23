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

package delete

import (
	"github.com/spf13/cobra"

	"github.com/ppc64le-cloud/pvsadm/cmd/delete/port"
	"github.com/ppc64le-cloud/pvsadm/pkg"
)

var Cmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the resources",
	Long:  `Delete the resources`,
}

func init() {
	Cmd.AddCommand(port.Cmd)
	Cmd.PersistentFlags().StringVarP(&pkg.Options.InstanceID, "instance-id", "i", "", "Instance ID of the PowerVS instance")
	_ = Cmd.MarkPersistentFlagRequired("instance-id")
}
