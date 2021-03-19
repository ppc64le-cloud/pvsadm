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
	"math/rand"
	"path/filepath"
	"time"

	"github.com/IBM-Cloud/bluemix-go/api/resource/resourcev1/management"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

const (
	ServiceType              = "cloud-object-storage"
	UseExistingPromptMessage = "Would You Like to use Available COS Instance for creating bucket?"
	CreatePromptMessage      = "Would you like to create new COS Instance?"
	ResourceGroupAPIRegion   = "global"
)

func createInstance(cli *client.Client) (string, error) {
	opt := pkg.ImageCMDOptions
	rand.Seed(time.Now().UnixNano())
	if opt.ResourceGrp == "" {
		resourceGroupQuery := management.ResourceGroupQuery{
			AccountID: cli.User.Account,
		}

		resGrpList, err := cli.ResGroupAPI.List(&resourceGroupQuery)
		if err != nil {
			return "", err
		}

		var resourceGroupNames []string
		for _, resgrp := range resGrpList {
			resourceGroupNames = append(resourceGroupNames, resgrp.Name)
		}
		opt.ResourceGrp = utils.SelectItem("Select ResourceGroup having required permissions for creating a service instance from the below:", resourceGroupNames)
	}

	if opt.NoPrompt {
		opt.InstanceName = fmt.Sprintf("pvsadm_cos_instance_%s_%d", time.Now().Format("20060102"), rand.Intn(100))
	} else {
		opt.InstanceName = utils.ReadUserInput("Type Name of the Cloud Object Storage instance:")
	}
	klog.Infof("Creating a cloud object storage instance: %s\n", opt.InstanceName)
	instanceID, err := cli.CreateServiceInstance(opt.InstanceName, ServiceType, opt.ServicePlan,
		opt.ResourceGrp, ResourceGroupAPIRegion)
	if err != nil {
		return "", err
	}
	return instanceID, nil
}

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
pvsadm image upload --bucket bucket1320 -f centos-8-latest.ova.gz --region <Region>

#If user likes to give different name to s3 Object
pvsadm image upload --bucket bucket1320 -f centos-8-latest.ova.gz -o centos8latest.ova.gz
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var s3Cli *client.S3Client
		var bucketExists bool = false
		var apikey string = pkg.Options.APIKey
		opt := pkg.ImageCMDOptions

		//Create bluemix client
		bxCli, err := client.NewClientWithEnv(apikey, pkg.Options.Environment, pkg.Options.Debug)

		if err != nil {
			return err
		}

		instances, err := bxCli.ListServiceInstances(ServiceType)
		if err != nil {
			return err
		}

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
				opt.InstanceName = instanceName
				klog.Infof("Found bucket %s in the %s instance", opt.BucketName, opt.InstanceName)
				break
			}
		}

		if len(instances) == 0 {
			klog.Infof("No active Cloud Object Storage instances were found in the account\n")
			if opt.NoPrompt || utils.AskConfirmation(CreatePromptMessage) {
				//Create a new instance
				instanceID, err := createInstance(bxCli)
				if err != nil {
					return err
				}
				klog.Infof("Cloud Object Storage InstanceID: %s\n", instanceID)
			}
			return fmt.Errorf("Create Cloud Object Storage instance either offline or use the pvsadm command\n")
		} else if opt.InstanceName == "" {
			klog.Infof("Bucket %s not found in the account provided\n", opt.BucketName)
			if !opt.NoPrompt && utils.AskConfirmation(UseExistingPromptMessage) {
				availableInstances := []string{}
				for name, _ := range instances {
					availableInstances = append(availableInstances, name)
				}
				selectedInstance := utils.SelectItem("Select Cloud Object Storage Instance:", availableInstances)
				opt.InstanceName = selectedInstance
				klog.Infof("Selected InstanceName is %s\n", opt.InstanceName)
			} else {
				//Ask If user wants to create new instance
				if opt.NoPrompt || utils.AskConfirmation(CreatePromptMessage) {
					//Create a new instance
					instanceID, err := createInstance(bxCli)
					if err != nil {
						return err
					}
					klog.Infof("Cloud Object Storage InstanceID: %s\n", instanceID)
				} else {
					return fmt.Errorf("Create Cloud Object Storage instance either offline or use the pvsadm command\n")
				}
			}
		}

		//create s3 client
		s3Cli, err = client.NewS3Client(bxCli, opt.InstanceName, opt.Region)
		if err != nil {
			return err
		}

		//Create a new bucket
		if !bucketExists {
			klog.Infof("Creating a new bucket %s\n", opt.BucketName)
			s3Cli, err = client.NewS3Client(bxCli, opt.InstanceName, opt.Region)
			if err != nil {
				return err
			}

			err = s3Cli.CreateBucket(opt.BucketName)
			if err != nil {
				return err
			}
		} else {
			//Check if object exists or not
			if opt.ObjectName == "" {
				opt.ObjectName = filepath.Base(opt.ImageName)
			}

			if s3Cli.CheckIfObjectExists(opt.BucketName, opt.ObjectName) {
				return fmt.Errorf("%s object already exists in the %s bucket", opt.ObjectName, opt.BucketName)
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
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.InstanceName, "cos-instance-name", "n", "", "Cloud Object Storage instance name.")
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.BucketName, "bucket", "b", "", "Cloud Object Storage bucket name.")
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.ImageName, "file", "f", "", "The PATH to the file to upload.")
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.ObjectName, "cos-object-name", "o", "", "Cloud Object Storage Object Name(Default: filename from --file|-f option)")
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.Region, "bucket-region", "r", "us-south", "Cloud Object Storage bucket region.")
	Cmd.PersistentFlags().BoolVar(&pkg.ImageCMDOptions.NoPrompt, "no-prompt", false, "option for running image upload e2e tests")
	_ = Cmd.MarkFlagRequired("bucket")
	_ = Cmd.MarkFlagRequired("file")
	Cmd.Flags().SortFlags = false
}
