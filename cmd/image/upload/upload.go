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

package upload

import (
	"fmt"
	"path/filepath"

	"github.com/IBM-Cloud/bluemix-go/api/resource/resourcev1/management"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

const (
	ServiceType              = "cloud-object-storage"
	UseExistingPromptMessage = "Would you like to use an existing COS Instance for creating bucket?"
	CreatePromptMessage      = "Would you like to create a new COS Instance?"
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
pvsadm image upload --bucket bucket0711 -f rhcos-461.ova.gz --instance-name pvsadm-cos-instance

#If user is planning to use available cos instance
pvsadm image upload  --bucket bucket0911 -f rhcos-461.ova.gz

#If user intents to create a new COS instance
pvsadm image upload --bucket bucket1320 -f centos-8-latest.ova.gz --resource-group <ResourceGroup_Name>

#if user is planning to create a bucket in particular region
pvsadm image upload --bucket bucket1320 -f centos-8-latest.ova.gz --bucket-region <Region>

#If user likes to give different name to s3 Object
pvsadm image upload --bucket bucket1320 -f centos-8-latest.ova.gz -o centos8latest.ova.gz

#upload using accesskey and secret key
pvsadm image upload --bucket bucket1320 -f centos-8-latest.ova.gz --bucket-region <Region> --accesskey <ACCESSKEY> --secretkey <SECRETKEY>
`,
	PreRunE: func(cmd *cobra.Command, args []string) error {

		if pkg.Options.Environment == "test" {
			return fmt.Errorf("image upload in test/staging env storage bucket is not supported")
		}

		// ensure that both, the AccessKey and SecretKey are set.
		if (len(pkg.ImageCMDOptions.AccessKey) > 0) != (len(pkg.ImageCMDOptions.SecretKey) > 0) {
			return fmt.Errorf("required both --accesskey and --secretkey values")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var s3Cli *client.S3Client
		var bucketExists bool
		var apikey string = pkg.Options.APIKey
		opt := pkg.ImageCMDOptions

		if opt.ObjectName == "" {
			opt.ObjectName = filepath.Base(opt.ImageName)
		}

		if pkg.ImageCMDOptions.AccessKey != "" && pkg.ImageCMDOptions.SecretKey != "" {
			s3Cli, err := client.NewS3ClientWithKeys(pkg.ImageCMDOptions.AccessKey, pkg.ImageCMDOptions.SecretKey, opt.Region)
			if err != nil {
				return err
			}
			//Check if object exists or not
			objectExists, err := s3Cli.CheckIfObjectExists(opt.BucketName, opt.ObjectName)
			if err != nil {
				return err
			}
			if objectExists {
				return fmt.Errorf("%s object already exists in the %s bucket", opt.ObjectName, opt.BucketName)
			}

			// upload the Image to S3 bucket
			return s3Cli.UploadObject(opt.ImageName, opt.ObjectName, opt.BucketName)

		}

		//Create bluemix client
		bxCli, err := client.NewClientWithEnv(apikey, pkg.Options.Environment, pkg.Options.Debug)

		if err != nil {
			return err
		}

		instances, err := bxCli.ListServiceInstances(ServiceType)
		if err != nil {
			return err
		}

		//check if bucket exists
		if opt.COSInstanceName != "" {
			s3Cli, err = client.NewS3Client(bxCli, opt.COSInstanceName, opt.Region)
			if err != nil {
				return err
			}

			bucketExists, err = s3Cli.CheckBucketExists(opt.BucketName)
			if err != nil {
				return err
			}

			if !bucketExists {
				klog.Infof("Bucket %s not found in the instance %s provided", opt.BucketName, opt.COSInstanceName)
			}

		} else if len(instances) != 0 {
			//check for bucket across the instances
			for instanceName := range instances {
				s3Cli, err = client.NewS3Client(bxCli, instanceName, opt.Region)
				if err != nil {
					return err
				}

				bucketExists, err = s3Cli.CheckBucketExists(opt.BucketName)
				if err != nil {
					return err
				}

				if bucketExists {
					opt.COSInstanceName = instanceName
					klog.Infof("Found bucket %s in the %s instance", opt.BucketName, opt.COSInstanceName)
					break
				}
			}
		} else if len(instances) == 0 {
			klog.Info("No active Cloud Object Storage instances were found in the account")
		}

		// Ask if user likes to use existing instance
		if opt.COSInstanceName == "" && len(instances) != 0 {
			klog.Infof("Bucket %s not found in the account provided", opt.BucketName)
			if utils.AskConfirmation(UseExistingPromptMessage) {
				availableInstances := []string{}
				for name := range instances {
					availableInstances = append(availableInstances, name)
				}
				selectedInstance, err := utils.SelectItem("Select Cloud Object Storage Instance:", availableInstances)
				if err != nil {
					return err
				}
				opt.COSInstanceName = selectedInstance
				klog.Infof("Selected InstanceName is %s", opt.COSInstanceName)
			}
		}

		//Create a new instance
		if opt.COSInstanceName == "" {
			if !utils.AskConfirmation(CreatePromptMessage) {
				return fmt.Errorf("create Cloud Object Storage instance either offline or use the pvsadm command")
			}
			if opt.ResourceGrp == "" {
				resourceGroupQuery := management.ResourceGroupQuery{
					AccountID: bxCli.User.Account,
				}

				resGrpList, err := bxCli.ResGroupAPI.List(&resourceGroupQuery)
				if err != nil {
					return err
				}

				var resourceGroupNames []string
				for _, resgrp := range resGrpList {
					resourceGroupNames = append(resourceGroupNames, resgrp.Name)
				}

				opt.ResourceGrp, err = utils.SelectItem("Select ResourceGroup having required permissions for creating a service instance from the below:", resourceGroupNames)
				if err != nil {
					return err
				}
			}

			opt.COSInstanceName = utils.ReadUserInput("Type Name of the Cloud Object Storage instance:")
			klog.Infof("Creating a new cos %s instance", opt.COSInstanceName)

			_, err = bxCli.CreateServiceInstance(opt.COSInstanceName, ServiceType, opt.ServicePlan,
				opt.ResourceGrp, ResourceGroupAPIRegion)
			if err != nil {
				return err
			}
		}

		//create s3 client
		s3Cli, err = client.NewS3Client(bxCli, opt.COSInstanceName, opt.Region)
		if err != nil {
			return err
		}
		if bucketExists {
			objectExists, err := s3Cli.CheckIfObjectExists(opt.BucketName, opt.ObjectName)
			if err != nil {
				return err
			}
			if objectExists {
				return fmt.Errorf("%s object already exists in the %s bucket", opt.ObjectName, opt.BucketName)
			}
		} else {
			//Create a new bucket
			klog.Infof("Creating a new bucket %s", opt.BucketName)
			s3Cli, err = client.NewS3Client(bxCli, opt.COSInstanceName, opt.Region)
			if err != nil {
				return err
			}

			err = s3Cli.CreateBucket(opt.BucketName)
			if err != nil {
				return err
			}
		}

		//upload the Image to S3 bucket
		err = s3Cli.UploadObject(opt.ImageName, opt.ObjectName, opt.BucketName)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	Cmd.Flags().StringVar(&pkg.ImageCMDOptions.ResourceGrp, "resource-group", "", "Name of user resource group.")
	Cmd.Flags().StringVar(&pkg.ImageCMDOptions.ServicePlan, "cos-serviceplan", "standard", "Cloud Object Storage Class type, available values are [standard, lite].")
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.COSInstanceName, "cos-instance-name", "n", "", "Cloud Object Storage instance name.")
	Cmd.Flags().MarkShorthandDeprecated("cos-instance-name", "please use --cos-instance-name")
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.BucketName, "bucket", "b", "", "Cloud Object Storage bucket name.")
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.ImageName, "file", "f", "", "The PATH to the file to upload.")
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.ObjectName, "cos-object-name", "o", "", "Cloud Object Storage Object Name(Default: filename from --file|-f option)")
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.Region, "bucket-region", "r", "us-south", "Cloud Object Storage bucket region.")
	Cmd.Flags().StringVar(&pkg.ImageCMDOptions.AccessKey, "accesskey", "", "Cloud Object Storage HMAC access key.")
	Cmd.Flags().StringVar(&pkg.ImageCMDOptions.SecretKey, "secretkey", "", "Cloud Object Storage HMAC secret key.")
	_ = Cmd.MarkFlagRequired("bucket")
	_ = Cmd.MarkFlagRequired("file")
	Cmd.Flags().SortFlags = false
}
