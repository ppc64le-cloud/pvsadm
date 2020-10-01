package events

import (
	"fmt"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"github.com/spf13/cobra"
	"time"
)

var (
	since time.Duration
)

var Cmd = &cobra.Command{
	Use:   "events",
	Short: "Get Powervs events",
	Long:  `Get the PowerVS events`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if pkg.Options.InstanceID == "" && pkg.Options.InstanceName == "" {
			return fmt.Errorf("--instance-name or --instance-name required")
		}
		return nil
	},
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
