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
	"os"
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
	s3Cli     SyncClient
	srcBucket string
	tgtBucket string
	srcObject string
}

// instance item for source and target instances
type InstanceItem struct {
	Source SyncClient
	Target []SyncClient
}

// sync constants
const (
	serviceType = "cloud-object-storage"
	maxWorkers  = 20
)

// Worker method to copy object from source bucket to target bucket
func copyWorker(copyJobs <-chan copyWorkload, results chan<- bool, workerId int) {
	for copyJob := range copyJobs {
		start := time.Now()
		klog.Infof("Copying object: %s src bucket: %s dest bucket: %s", copyJob.srcObject, copyJob.srcBucket, copyJob.tgtBucket)
		err := copyJob.s3Cli.CopyObjectToBucket(copyJob.srcBucket, copyJob.tgtBucket, copyJob.srcObject)
		if err != nil {
			klog.Errorf("copy object %s failed, err: %v", copyJob.srcObject, err)
			results <- false
		}
		duration := time.Since(start)
		klog.Infof("Copying object: %s from bucket: %s to bucket: %s took %v", copyJob.srcObject, copyJob.srcBucket, copyJob.tgtBucket, duration)
		results <- true
	}
}

// Method to create the list of required instances
func createInstanceList(spec []pkg.Spec, bxCli *client.Client) ([]InstanceItem, error) {
	var instanceList []InstanceItem
	for _, item := range spec {
		instance := InstanceItem{}
		s3Cli, err := NewS3Client(bxCli, item.Source.Cos, item.Source.Region)
		if err != nil {
			return nil, err
		}

		instance.Source = s3Cli
		for _, targetItem := range item.Target {
			s3Cli, err := NewS3Client(bxCli, item.Source.Cos, targetItem.Region)
			if err != nil {
				return nil, err
			}
			instance.Target = append(instance.Target, s3Cli)
		}
		instanceList = append(instanceList, instance)
	}
	return instanceList, nil
}

// Method to calculate channels
func calculateChannels(spec []pkg.Spec, instanceList []InstanceItem) (int, error) {
	totalChannels := 0
	for item_no, item := range spec {
		_, err := instanceList[item_no].Source.CheckBucketLocationConstraint(item.Source.Bucket, item.Source.Region+"-"+item.Source.StorageClass)
		if err != nil {
			klog.Errorf("location constraint verification failed for src bucket %s", item.Source.Bucket)
			return 0, err
		}

		selectedObjects, err := instanceList[item_no].Source.SelectObjects(item.Source.Bucket, item.Source.Object)
		if err != nil {
			klog.Errorf("select Objects failed, err: %v", err)
			return 0, err
		}

		numTargets := len(item.Target)
		totalChannelsForSrc := numTargets * len(selectedObjects)
		totalChannels = totalChannels + totalChannelsForSrc
	}
	return totalChannels, nil
}

// Method to select and copy objects
func copyObjects(spec []pkg.Spec, instanceList []InstanceItem, copyJobs chan<- copyWorkload) error {
	for item_no, item := range spec {
		selectedObjects, err := instanceList[item_no].Source.SelectObjects(item.Source.Bucket, item.Source.Object)
		if err != nil {
			klog.Errorf("select Objects failed, err: %v", err)
			return err
		}

		klog.Infof("Selected Objects from bucket %s: %s", item.Source.Bucket, strings.Join(selectedObjects, ", "))
		for targetItemNo, targetItem := range item.Target {
			_, err = instanceList[item_no].Target[targetItemNo].CheckBucketLocationConstraint(targetItem.Bucket, targetItem.Region+"-"+targetItem.StorageClass)
			if err != nil {
				klog.Errorf("location constraint verification failed for dest bucket %s", targetItem.Bucket)
				return errors.New("bucket location constraint verification failed")
			}

			for _, srcObject := range selectedObjects {
				copyJob := copyWorkload{
					s3Cli:     instanceList[item_no].Target[targetItemNo],
					srcBucket: item.Source.Bucket,
					tgtBucket: targetItem.Bucket,
					srcObject: srcObject,
				}
				copyJobs <- copyJob
			}
		}
	}

	return nil
}

// Method to get the results from channels
func getResults(results <-chan bool, totalChannels int) bool {
	passedCopies := 0
	failedCopies := 0
	for i := 1; i <= totalChannels; i++ {
		if !<-results {
			failedCopies = failedCopies + 1
			continue
		}
		passedCopies = passedCopies + 1
	}
	klog.Infof("No of copies passed: %d No of copies failed: %d", passedCopies, failedCopies)
	return failedCopies == 0
}

// Method to get specifications
func getSpec(specfileName string) ([]pkg.Spec, error) {
	var spec []pkg.Spec
	// Unmashalling yaml file
	yamlFile, err := os.ReadFile(specfileName)
	if err != nil {
		klog.Errorf("read yaml failed, err: %v", err)
		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, &spec)
	if err != nil {
		klog.Errorf("unmarshal failed, err: %v", err)
		return nil, err
	}

	return spec, nil
}

// Method sync objects
func syncObjects(spec []pkg.Spec, instanceList []InstanceItem) error {
	// Calculating total channels
	totalChannels, err := calculateChannels(spec, instanceList)
	if err != nil {
		return err
	}

	// Creating workers and channels
	copyJobs := make(chan copyWorkload, totalChannels)
	results := make(chan bool, totalChannels)
	for worker := 1; worker <= maxWorkers; worker++ {
		go copyWorker(copyJobs, results, worker)
	}

	// Copy objects
	err = copyObjects(spec, instanceList, copyJobs)
	if err != nil {
		return err
	}
	close(copyJobs)

	// Wait and get results from channels
	res := getResults(results, totalChannels)
	if !res {
		return errors.New("copy objects failed")
	}

	return nil
}

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

		var apikey string = pkg.Options.APIKey
		opt := pkg.ImageCMDOptions
		start := time.Now()

		//Create bluemix client
		bxCli, err := client.NewClientWithEnv(apikey, pkg.Options.Environment, pkg.Options.Debug)
		if err != nil {
			return err
		}

		// Generate Specifications
		spec, err := getSpec(opt.SpecYAML)
		if err != nil {
			return err
		}

		// Create necessary objects
		instanceList, err := createInstanceList(spec, bxCli)
		if err != nil {
			return err
		}

		// Sync Objects
		err = syncObjects(spec, instanceList)
		if err != nil {
			return err
		}

		// Calculate total elapsed time
		duration := time.Since(start)
		klog.Infof("Total elapsed time: %v", duration)
		return nil
	},
}

// Init method
func init() {
	Cmd.Flags().StringVarP(&pkg.ImageCMDOptions.SpecYAML, "spec-file", "s", "", "The PATH to the spec file to be used")
	_ = Cmd.MarkFlagRequired("spec-file")
	Cmd.Flags().SortFlags = false
}
