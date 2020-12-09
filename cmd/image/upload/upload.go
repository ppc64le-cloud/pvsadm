package upload

import (
	"fmt"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/ppc64le-cloud/pvsadm/pkg/client/s3utils"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

const (
	ServiceType              = "cloud-object-storage"
	UseExistingPromptMessage = "Would You Like to use Available COS Instance for creating bucket?"
	CreatePromptMessage      = "Would you like to create new COS Instance?"
	ResourceGroupAPIRegion   = "global"
)

var Cmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload the image to the IBM COS",
	Long: `Upload the image to the IBM COS
pvsadm image upload --help for information

# Set the API key or feed the --api-key commandline argument
export IBMCLOUD_API_KEY=<IBM_CLOUD_API_KEY>

Examples:

# using InstanceName
pvsadm image upload --bucket bucket0711 -o rhcos-461.ova.gz --cos-instance-name pvsadm-cos-instance

#If user is planning to use available cos instance
pvsadm image upload  --bucket bucket0911 -o rhcos-461.ova.gz

#If user intents to create a new COS instance
pvsadm image upload --bucket basheerbucket1320 -o centos-8-latest.ova.gz --resource-group <ResourceGroup_Name>

#if user is planning to create a bucket in particular region
pvsadm image upload --bucket basheerbucket1320 -o centos-8-latest.ova.gz --region <Region>
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var s3Cli *client.S3Client
		apikey := pkg.Options.APIKey
		opt := pkg.ImageCMDOptions
		bucketExists := false

		//Create bluemix client
		bxCli, err := client.NewClient(apikey)
		if err != nil {
			return err
		}
		instances, err := s3utils.GetInstances(bxCli, ServiceType)
		if err != nil {
			return err
		}

		instanceExists := len(instances) != 0

		if opt.InstanceName != "" {
			s3Cli, err = client.NewS3Client(bxCli, opt.InstanceName, opt.Region)
			if err != nil {
				return err
			}
			bucketExists, err = s3Cli.CheckBucketExists(opt.BucketName)
			if err != nil {
				return err
			}
		} else if instanceExists {
			//check for bucket across the instances
			for instanceName, _ := range instances {
				s3Cli, err = client.NewS3Client(bxCli, instanceName, opt.Region)
				if err != nil {
					return err
				}
				bucketExists, err = s3Cli.CheckBucketExists(opt.BucketName)
				if err != nil {
					return err
				}
				if bucketExists {
					klog.Infof("Found bucket %s in the %s instance\n", opt.BucketName, instanceName)
					break
				}
			}

			if !bucketExists {
				klog.Infof("bucket %s not found in the account provided\n", opt.BucketName)
				//if bucket doesn't exists,ask user if he wants to use existing cos instance
				if s3utils.AskYesOrNo(UseExistingPromptMessage, 3) {
					//List of Available COS instances
					klog.Infof("Select a COS Instance\n")
					instanceNames := []string{}
					for name, _ := range instances {
						instanceNames = append(instanceNames, name)
					}
					count := 0
					for _, name := range instanceNames {
						fmt.Printf("%d. %s (%s)\n", count, name, instances[name])
						count = count + 1
					}
					input := s3utils.SelectCosInstance(len(instanceNames), 3)
					if input == -1 {
						return fmt.Errorf("Please select a valid COS Instance\n")
					}
					opt.InstanceName = instanceNames[input]
					klog.Infof("Selected InstanceName is %s\n", opt.InstanceName)
					s3Cli, err = client.NewS3Client(bxCli, opt.InstanceName, opt.Region)
				} else {
					if s3utils.AskYesOrNo(CreatePromptMessage, 3) {
						name := s3utils.ReadInstanceNameFromUser()
						klog.Infof("Creating a new cos %s instance\n", name)
						_, err = client.CreateServiceInstance(bxCli.Session, name, ServiceType, opt.ServicePlan,
							opt.ResourceGrp, ResourceGroupAPIRegion)
						if err != nil {
							return err
						}
						s3Cli, err = client.NewS3Client(bxCli, name, opt.Region)
						if err != nil {
							return err
						}
					} else {
						return fmt.Errorf("please create cos instance either offline or use the pvsadm command\n")
					}
				}
			}
		} else {
			name := s3utils.ReadInstanceNameFromUser()
			klog.Infof("Creating a new cos %s instance\n", name)
			_, err = client.CreateServiceInstance(bxCli.Session, name, ServiceType, opt.ServicePlan,
				opt.ResourceGrp, ResourceGroupAPIRegion)
			if err != nil {
				return err
			}
			s3Cli, err = client.NewS3Client(bxCli, name, opt.Region)
			if err != nil {
				return err
			}
		}

		//Create a new bucket
		if !bucketExists {
			klog.Infof("Creating a new bucket %s\n", opt.BucketName)
			err = s3Cli.CreateBucket(opt.BucketName)
			if err != nil {
				return err
			}
		}
		//upload the Image to S3 bucket
		err = s3Cli.UploadObject(opt.ImageName, opt.BucketName)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	Cmd.Flags().StringVar(&pkg.ImageCMDOptions.ResourceGrp, "resource-group", "default", "Provide Resource-Group")
	Cmd.Flags().StringVar(&pkg.ImageCMDOptions.ServicePlan, "service-plan", "standard", "Provide serviceplan type")
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.InstanceName, "cos-instance-name", "n", "", "Name of the COS instance")
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.BucketName, "bucket", "b", "", "cloud-object-storage bucket name")
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.ImageName, "object-name", "o", "", "S3 object name to be uploaded")
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.Region, "region", "r", "us-south", "COS bucket location")
	_ = Cmd.MarkFlagRequired("bucket")
	_ = Cmd.MarkFlagRequired("object-name")
	Cmd.Flags().SortFlags = false
}
