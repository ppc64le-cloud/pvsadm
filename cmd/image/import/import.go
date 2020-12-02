package _import

import (
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"os"
	"strings"
)

var Cmd = &cobra.Command{
	Use:   "import",
	Short: "Import the image into PowerVS instances",
	Long: `Import the image into PowerVS instances
pvsadm image import --help for information

# Set the API key or feed the --api-key commandline argument
export IBMCLOUD_API_KEY=<IBM_CLOUD_API_KEY>

Examples:

# importing image from default region and using default storage type
pvsadm image import -n upstream-core-lon04 -b <BUCKETNAME> --accesskey <ACCESSKEY> --secretkey <SECRETKEY> --object-name rhel-83-10032020.ova.gz --image-name test-image

# with user provided storage type and region
pvsadm image import -n upstream-core-lon04 -b <BUCKETNAME> --accesskey <ACCESSKEY> --secretkey <SECRETKEY> -r <REGION> --storagetype <STORAGETYPE> --object-name rhel-83-10032020.ova.gz --image-name test-image

# If user wants to specify the type of OS
pvsadm image import -n upstream-core-lon04 -b <BUCKETNAME> --accesskey <ACCESSKEY> --secretkey <SECRETKEY> --object-name rhel-83-10032020.ova.gz --image-name test-image --ostype <OSTYPE>
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opt := pkg.ImageCMDOptions
		//validate inputs
		validOsType := []string{"aix", "ibmi", "redhat", "sles"}
		validStorageType := []string{"tier3", "tier1"}

		if opt.OsType != "" && !utils.Contains(validOsType, strings.ToLower(opt.OsType)) {
			klog.Errorf("Provide valid OsType.. allowable values are [aix, ibmi, redhat, sles]")
			os.Exit(1)
		}

		if !utils.Contains(validStorageType, strings.ToLower(opt.StorageType)) {
			klog.Errorf("Provide valid StorageType.. allowable values are [tier1, tier3]")
			os.Exit(1)
		}

		c, err := client.NewClient(pkg.Options.APIKey)
		if err != nil {
			return err
		}

		pvmclient, err := client.NewPVMClient(c, opt.InstanceID, opt.InstanceName)
		if err != nil {
			return err
		}

		image, err := pvmclient.ImgClient.ImportImage(pvmclient.InstanceID, opt.ImageName, opt.ImageFilename, opt.Region,
			opt.AccessKey, opt.SecretKey, opt.BucketName, strings.ToLower(opt.OsType), strings.ToLower(opt.StorageType))
		if err != nil {
			return err
		}

		klog.Infof("Importing Image %s is currently in %s state, Please check the Progress in the IBM Cloud UI\n", *image.Name, image.State)
		return nil
	},
}

func init() {
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.InstanceName, "instance-name", "n", "", "Instance name of the PowerVS")
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.InstanceID, "instance-id", "i", "", "Instance ID of the PowerVS instance")
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.BucketName, "bucket", "b", "", "Cloud Storage bucket name")
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.Region, "region", "r", "", "Cloud Storage Region")
	Cmd.Flags().StringVar(&pkg.ImageCMDOptions.AccessKey, "accesskey", "", "Cloud Storage access key; required for import image")
	Cmd.Flags().StringVar(&pkg.ImageCMDOptions.SecretKey, "secretkey", "", "Cloud Storage secret key; required for import image")
	Cmd.Flags().StringVar(&pkg.ImageCMDOptions.ImageName, "image-name", "", "Name to give imported image")
	Cmd.Flags().StringVar(&pkg.ImageCMDOptions.ImageFilename, "object-name", "", "Cloud Storage image filename")
	Cmd.Flags().StringVar(&pkg.ImageCMDOptions.OsType, "ostype", "", "Image OS Type, accepted values are[aix, ibmi, redhat, sles]")
	Cmd.Flags().StringVar(&pkg.ImageCMDOptions.StorageType, "storagetype", "tier3", "Storage type, accepted values are [tier1, tier3]")
	_ = Cmd.MarkFlagRequired("bucket")
	_ = Cmd.MarkFlagRequired("accesskey")
	_ = Cmd.MarkFlagRequired("secretkey")
	_ = Cmd.MarkFlagRequired("image-name")
	_ = Cmd.MarkFlagRequired("object-name")
	_ = Cmd.MarkFlagRequired("region")
}
