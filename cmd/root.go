package cmd

import (
	"github.com/ppc64le-cloud/pvsadm/cmd/get"
	"github.com/ppc64le-cloud/pvsadm/cmd/purge"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/audit"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "pvsadm",
	Short: "pvsadm is a command for managing powervs infra",
	Long: `Power Systems Virtual Server projects deliver flexible compute capacity for Power Systems workloads.
Integrated with the IBM Cloud platform for on-demand provisioning.

This is a tool built for the Power Systems Virtual Server helps managing and maintaining the resources easily`,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(purge.Cmd)
	rootCmd.AddCommand(get.Cmd)
	rootCmd.PersistentFlags().StringVarP(&pkg.Options.APIKey, "api-key", "k", "", "IBMCLOUD API Key(env name: IBMCLOUD_API_KEY)")
	rootCmd.PersistentFlags().BoolVar(&pkg.Options.Debug, "debug", false, "Enable PowerVS debug option")
	rootCmd.PersistentFlags().StringVar(&pkg.Options.AuditFile, "audit-file", "pvsadm.log", "Audit logs for the tool")
	rootCmd.Flags().SortFlags = false
	rootCmd.PersistentFlags().SortFlags = false
	audit.Logger = audit.New(pkg.Options.AuditFile)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		klog.Errorln(err)
		os.Exit(1)
	}
}
