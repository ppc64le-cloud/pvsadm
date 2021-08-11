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
	"errors"
	"io/ioutil"
	"strings"
	"time"

	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

// copy workload type for copy worker method
type copyWorkload struct {
	s3Cli     client.S3Client
	srcBucket string
	tgtBucket string
	srcObject string
}

// sync constants
const (
	ServiceType     = "cloud-object-storage"
	NoOfCopyWorkers = 20
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
    bucket: bucket-fqsdgh
    cos: cos-test-buphij
    object: ".txt"
    storageClass: smart
    region: jp-tok
  target:
  - bucket: bucket-ielfqq
    storageClass: smart
    region: jp-tok
  - bucket: bucket-kkoyrg
    storageClass: smart
    region: jp-tok
- source:
    bucket: bucket-embocx
    cos: cos-test-icrbul
    object: ""
    storageClass: smart
    region: us-east
  target:
  - bucket: bucket-xgskog
    storageClass: standard
    region: jp-tok
  - bucket: bucket-nomoer
    storageClass: cold
    region: jp-tok

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var s3Cli *client.S3Client
		var selectedObjects []string

		var apikey string = pkg.Options.APIKey
		opt := pkg.ImageCMDOptions
		start := time.Now()

		//Create bluemix client
		bxCli, err := client.NewClientWithEnv(apikey, pkg.Options.Environment, pkg.Options.Debug)
		if err != nil {
			return err
		}

		// Unmashalling yaml file
		var spec []pkg.Spec
		yamlFile, err := ioutil.ReadFile(opt.SpecYAML)
		if err != nil {
			klog.Errorf("ERROR: Read yaml failed : %v", err)
			return err
		}

		err = yaml.Unmarshal(yamlFile, &spec)
		if err != nil {
			klog.Errorf("Unmarshal: %v", err)
			return err
		}

		copyWorker := func(copyJobs <-chan copyWorkload, results chan<- bool, workerId int) {
			for copyJob := range copyJobs {
				start := time.Now()
				klog.Infof("Copying object: %s src bucket: %s dest bucket: %s", copyJob.srcObject, copyJob.srcBucket, copyJob.tgtBucket)
				err = copyJob.s3Cli.CopyObjectToBucket(copyJob.srcBucket, copyJob.tgtBucket, copyJob.srcObject)
				if err != nil {
					klog.Errorf("ERROR: %v, Copy object %s failed", err, copyJob.srcObject)
					results <- false
				}
				duration := time.Since(start)
				klog.Infof("Copying object: %s from bucket: %s to bucket: %s took %v", copyJob.srcObject, copyJob.srcBucket, copyJob.tgtBucket, duration)
				results <- true
			}
		}

		// Calculating total channels required
		totalChannels := 0
		for _, item := range spec {
			s3Cli, err = client.NewS3Client(bxCli, item.Source.Cos, item.Source.Region)
			if err != nil {
				return err
			}

			_, err = s3Cli.CheckBucketLocationConstraint(item.Source.Bucket, item.Source.Region+"-"+item.Source.StorageClass)
			if err != nil {
				klog.Errorf("Location constraint verification failed for src bucket: %s", item.Source.Bucket)
				return err
			}

			selectedObjects, err = s3Cli.SelectObjects(item.Source.Bucket, item.Source.Object)
			if err != nil {
				klog.Errorf("Select Objects failed: %v", err)
				return err
			}

			noOfTargets := len(item.Target)
			totalChannelsForSrc := noOfTargets * len(selectedObjects)
			totalChannels = totalChannels + totalChannelsForSrc
		}

		// Creating workers and channels
		copyJobs := make(chan copyWorkload, totalChannels)
		results := make(chan bool, totalChannels)
		for w := 1; w <= NoOfCopyWorkers; w++ {
			go copyWorker(copyJobs, results, w)
		}

		for _, item := range spec {
			// Creating S3 client
			s3Cli, err = client.NewS3Client(bxCli, item.Source.Cos, item.Source.Region)
			if err != nil {
				return err
			}

			selectedObjects, err = s3Cli.SelectObjects(item.Source.Bucket, item.Source.Object)
			if err != nil {
				klog.Errorf("Select Objects failed: %v", err)
				return err
			}

			klog.Infof("Selected Objects from bucket %s: %s", item.Source.Bucket, strings.Join(selectedObjects, ", "))
			for _, targetItem := range item.Target {
				s3Cli, err = client.NewS3Client(bxCli, item.Source.Cos, targetItem.Region)
				if err != nil {
					return err
				}

				_, err = s3Cli.CheckBucketLocationConstraint(targetItem.Bucket, targetItem.Region+"-"+targetItem.StorageClass)
				if err != nil {
					klog.Errorf("Location constraint verification failed for dest bucket: %s", targetItem.Bucket)
					return errors.New("bucket location constraint verification failed")
				}

				for _, srcObject := range selectedObjects {
					copyJob := copyWorkload{
						s3Cli:     *s3Cli,
						srcBucket: item.Source.Bucket,
						tgtBucket: targetItem.Bucket,
						srcObject: srcObject,
					}
					copyJobs <- copyJob
				}

			}
		}

		passedCopies := 0
		failedCopies := 0
		for i := 1; i <= totalChannels; i++ {
			if !<-results {
				failedCopies = failedCopies + 1
				continue
			}
			passedCopies = passedCopies + 1
		}
		close(copyJobs)

		duration := time.Since(start)
		klog.Infof("No of copies passed: %d No of copies failed: %d Total elapsed time: %v", passedCopies, failedCopies, duration)
		if failedCopies > 0 {
			return errors.New("copy objects failed")
		}
		return nil
	},
}

func init() {
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.SpecYAML, "spec-file", "s", "", "The PATH to the spec file to be used")
	_ = Cmd.MarkFlagRequired("spec-file")
	Cmd.Flags().SortFlags = false
}
