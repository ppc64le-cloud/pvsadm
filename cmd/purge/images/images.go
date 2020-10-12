package images

import (
	"fmt"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/audit"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

const deletePromptMessage = "Deleting all the above images, images can't be claimed back once deleted. Do you really want to continue?"

var Cmd = &cobra.Command{
	Use:   "images",
	Short: "Purge the powervs images",
	Long: `Purge the powervs images!
pvsadm purge --help for information
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		klog.Infof("Purge the images for the instance: %v", pkg.Options.InstanceID)
		opt := pkg.Options

		c, err := client.NewClient(opt.APIKey)
		if err != nil {
			return err
		}

		pvmclient, err := client.NewPVMClient(c, opt.InstanceID, opt.InstanceName)
		if err != nil {
			return err
		}

		images, err := pvmclient.ImgClient.GetAllPurgeable(opt.Before, opt.Since, opt.Expr)
		if err != nil {
			return fmt.Errorf("failed to get the list of images: %v", err)
		}
		table := utils.NewTable()

		table.Render(images, []string{"href", "specifications"})
		if !opt.DryRun && len(images) != 0 {
			if opt.NoPrompt || utils.AskYesOrNo(deletePromptMessage) {
				for _, image := range images {
					klog.Infof("Deleting the %s, and ID: %s", *image.Name, *image.ImageID)
					err = pvmclient.ImgClient.Delete(*image.ImageID)
					if err != nil {
						if opt.IgnoreErrors {
							klog.Infof("error occurred while deleting the image: %v", err)
						} else {
							return err
						}
					}
					audit.Log("images", "delete", pvmclient.InstanceName+":"+*image.Name)
				}
			}
		}
		return nil
	},
}
