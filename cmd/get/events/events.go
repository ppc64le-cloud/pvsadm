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

package events

import (
	"time"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
)

var (
	since time.Duration
)

var Cmd = &cobra.Command{
	Use:   "events",
	Short: "Get Powervs events",
	Long:  `Get the PowerVS events`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return utils.EnsurePrerequisitesAreSet(pkg.Options.APIKey, pkg.Options.WorkspaceID, pkg.Options.WorkspaceName)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		opt := pkg.Options

		c, err := client.NewClientWithEnv(opt.APIKey, opt.Environment, opt.Debug)
		if err != nil {
			klog.Errorf("failed to create a session with IBM cloud, err: %v", err)
			return err
		}

		pvmclient, err := client.NewPVMClientWithEnv(c, opt.WorkspaceID, opt.WorkspaceName, opt.Environment)
		if err != nil {
			return err
		}
		events, err := pvmclient.EventsClient.GetPcloudEventsGetsince(since)
		if err != nil {
			return err
		}
		table := utils.NewTable()
		table.Render(events.Payload.Events, []string{"user", "timestamp"})

		return nil
	},
}

func init() {
	Cmd.PersistentFlags().DurationVar(&since, "since", 24*time.Hour, "Show events since")
}
