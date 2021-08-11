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

package sync

import (
	"fmt"
	"io/ioutil"

	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

type Spec []struct {
	Source struct {
		Bucket string `yaml:"bucket"`
		Cos    string `yaml:"cos"`
		Object string `yaml:"object"`
		Plan   string `yaml:"plan"`
		Region string `yaml:"region"`
	} `yaml:"source"`
	Target []struct {
		Bucket string `yaml:"bucket"`
		Plan   string `yaml:"plan"`
		Region string `yaml:"region"`
	} `yaml:"target"`
}

const (
	ServiceType = "cloud-object-storage"
)

var Cmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync images between IBM COS buckets",
	Long: `Sync images between IBM COS buckets
pvsadm image sync --help for information

# Set the API key or feed the --api-key commandline argument
export IBMCLOUD_API_KEY=<IBM_CLOUD_API_KEY>

Examples:

# using spec yaml file
pvsadm image sync --spec-file spec.yaml

Sample spec.yaml file:
---
- source:
    bucket: rhcos-bucket-us-south
    cos: cos-service-name-1
    object: regex
    plan: smart
    region: us-south
  target:
    - bucket: rhcos-bucket-us-east
      plan: smart
      region: us-east
    - bucket: rhcos-bucket-london
      plan: standard
      region: london
    - bucket: rhcos-bucket-london-lite
      plan: lite
      region: london
- source:
    bucket: rhcos-bucket-us-south
    cos: cos-service-name-2
    object: regex
    plan: smart
    region: us-south
  target:
    - bucket: rhcos-bucket-us-east
      plan: smart
      region: us-east
    - bucket: rhcos-bucket-london
      plan: standard
      region: london
    - bucket: rhcos-bucket-london-lite
      plan: lite
      region: london

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var s3Cli *client.S3Client
		var apikey string = pkg.Options.APIKey
		opt := pkg.ImageCMDOptions

		//Create bluemix client
		bxCli, err := client.NewClientWithEnv(apikey, pkg.Options.Environment, pkg.Options.Debug)

		if err != nil {
			return err
		}

		// Unmashalling yaml file
		var spec Spec
		yamlFile, err := ioutil.ReadFile(opt.SpecYAML)
		if err != nil {
			klog.Errorf("yamlFile.Get err   #%v ", err)
			return err
		}
		err = yaml.Unmarshal(yamlFile, &spec)
		if err != nil {
			klog.Errorf("Unmarshal: %v", err)
			return err
		}

		var selectedObjects []string
		var locationRes bool

		for _, item := range spec {

			// Creating S3 client
			s3Cli, err = client.NewS3Client(bxCli, item.Source.Cos, item.Source.Region)
			if err != nil {

				return err
			}
			locationRes, _ = s3Cli.CheckBucketLocationConstraint(item.Source.Bucket, item.Source.Region+"-"+item.Source.Plan)
			if !locationRes {
				klog.Errorf("Location constraint verification failed for src bucket: %s", item.Source.Bucket)
				continue
			}
			selectedObjects, err = s3Cli.SelectObjects(item.Source.Bucket, item.Source.Object)
			if err != nil {
				klog.Errorf("Select error: %v", err)
				return err
			}

			fmt.Println("Selected Objects:", selectedObjects)

			for _, targetItem := range item.Target {
				s3Cli, err = client.NewS3Client(bxCli, item.Source.Cos, targetItem.Region)
				if err != nil {

					return err
				}
				locationRes, _ = s3Cli.CheckBucketLocationConstraint(targetItem.Bucket, targetItem.Region+"-"+targetItem.Plan)
				if !locationRes {
					klog.Errorf("Location constraint verification failed for dest bucket: %s", targetItem.Bucket)
					continue
				}
				for _, srcObject := range selectedObjects {
					err = s3Cli.CopyObjectToBucket(item.Source.Bucket, targetItem.Bucket, srcObject)
					if err != nil {

						return err
					}
					klog.Infof("Copy successful src bucket: %s dest bucket: %s object: %s", item.Source.Bucket, targetItem.Bucket, srcObject)
				}
			}

		}

		return nil
	},
}

func init() {
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.SpecYAML, "spec-file", "s", "", "The PATH to the spec file to be used")
	_ = Cmd.MarkFlagRequired("spec-file")
	Cmd.Flags().SortFlags = false
}
