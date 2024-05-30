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

package get

import (
	"github.com/spf13/cobra"

	"github.com/ppc64le-cloud/pvsadm/cmd/get/cloudconnections"
	"github.com/ppc64le-cloud/pvsadm/cmd/get/events"
	"github.com/ppc64le-cloud/pvsadm/cmd/get/peravailability"
	"github.com/ppc64le-cloud/pvsadm/cmd/get/ports"
	"github.com/ppc64le-cloud/pvsadm/pkg"
)

var Cmd = &cobra.Command{
	Use:     "get",
	Short:   "Get the resources",
	Long:    `Get the resources`,
	GroupID: "resource",
}

func init() {
	Cmd.AddCommand(cloudconnections.Cmd)
	Cmd.AddCommand(events.Cmd)
	Cmd.AddCommand(peravailability.Cmd)
	Cmd.AddCommand(ports.Cmd)
	Cmd.PersistentFlags().StringVarP(&pkg.Options.WorkspaceID, "instance-id", "i", "", "Instance ID of the PowerVS instance")
	Cmd.PersistentFlags().MarkDeprecated("instance-id", "instance-id is deprecated, workspace-id should be used")
	Cmd.PersistentFlags().StringVarP(&pkg.Options.WorkspaceName, "instance-name", "n", "", "Instance name of the PowerVS")
	Cmd.PersistentFlags().MarkDeprecated("instance-name", "instance-name is deprecated, workspace-name should be used")
	Cmd.PersistentFlags().StringVarP(&pkg.Options.WorkspaceID, "workspace-id", "", "", "Workspace ID of the PowerVS instance")
	Cmd.PersistentFlags().StringVarP(&pkg.Options.WorkspaceName, "workspace-name", "", "", "Workspace name of the PowerVS")
}
