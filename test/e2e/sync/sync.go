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
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/IBM-Cloud/bluemix-go/api/resource/resourcev1/management"
	"github.com/IBM-Cloud/bluemix-go/api/resource/resourcev2/controllerv2"
	"github.com/IBM/ibm-cos-sdk-go/aws"
	"github.com/IBM/ibm-cos-sdk-go/service/s3"
	"github.com/ppc64le-cloud/pvsadm/pkg"
	"github.com/ppc64le-cloud/pvsadm/pkg/client"
	"github.com/ppc64le-cloud/pvsadm/pkg/utils"
	"github.com/ppc64le-cloud/pvsadm/test/e2e/framework"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// Test case variables
var (
	err               error
	bxCli             *client.Client
	APIKey            = os.Getenv("IBMCLOUD_API_KEY")
	ObjectsFolderName = "tempFolder"
	SpecFileName      = "spec/spec.yaml"
)

// Test case constants
const (
	serviceType            = "cloud-object-storage"
	resourceGroupAPIRegion = "global"
	servicePlan            = "standard"
	debug                  = false
	recursive              = false
	typeServiceInstance    = "service_instance"
)

// Test configurations
var (
	numSources          = 2
	numTargetsPerSource = 2
	numObjects          = 200
	numUploadWorkers    = 20
)

// Run sync command
func runSyncCMD(args ...string) (int, string, string) {
	args = append([]string{"image", "sync"}, args...)
	return utils.RunCMD("pvsadm", args...)
}

// Create Specifications yaml file
func createSpecFile(spec []pkg.Spec) error {
	klog.V(4).Info("STEP: Creating Spec file")
	dir, err := os.MkdirTemp(".", "spec")
	if err != nil {
		klog.Errorf("unable to create temporary directory, err: %v", err)
		return err
	}

	file, err := os.CreateTemp(dir, "spec.*.yaml")
	if err != nil {
		klog.Errorf("unable to create a temp file, err: %v", err)
		return err
	}
	defer file.Close()

	SpecFileName = file.Name()
	specString, merr := yaml.Marshal(&spec)
	if merr != nil {
		klog.Errorf("marshal failed, err: %v", merr)
		return merr
	}

	_, err = file.WriteString(string(specString))
	if err != nil {
		klog.Errorf("error while writing in the file, err: %v", err)
		return err
	}

	klog.V(3).Info("Specifications for e2e sync test", string(specString))
	return nil
}

// Create Cloud Object Storage Service instance
func createCOSInstance(instanceName string) error {
	klog.V(4).Infof("STEP: Creating COS instance : %s", instanceName)

	resourceGroupQuery := management.ResourceGroupQuery{
		AccountID: bxCli.User.Account,
	}
	resGrpList, err := bxCli.ResGroupAPI.List(&resourceGroupQuery)
	if err != nil {
		klog.Errorf("error while listing resource groups, err: %v", err)
		return err
	}

	var resourceGroupNames []string
	for _, resgrp := range resGrpList {
		resourceGroupNames = append(resourceGroupNames, resgrp.Name)
	}
	klog.V(3).Infof("Resource Group names: %v", resourceGroupNames)

	_, err = bxCli.CreateServiceInstance(instanceName, serviceType, servicePlan,
		resourceGroupNames[0], resourceGroupAPIRegion)
	if err != nil {
		klog.Errorf("unable to create Service Instance, err: %v", err)
		return err
	}

	return nil
}

// Delete Cloud Object Storage Service instance
func deleteCOSInstance(instanceName string) error {
	klog.V(4).Infof("STEP: Deleting COS instance %s", instanceName)
	svcs, err := bxCli.ResourceClientV2.ListInstances(controllerv2.ServiceInstanceQuery{
		Type: typeServiceInstance,
		Name: instanceName,
	})
	if err != nil {
		klog.Errorf("unable to list instance, err: %v", err)
		return err
	}

	for _, svc := range svcs {
		if svc.Name == instanceName {
			err = bxCli.DeleteServiceInstance(svc.ID, recursive)
			if err != nil {
				klog.Errorf("unable to delete Service Instance, err: %v", err)
				return err
			}
			klog.V(2).Infof("Service Instance Deleted: %s", svc.Name)
		}
	}

	return nil
}

// Create S3 bucket in the given region and storage class
func createBucket(bucketName string, cos string, region string, storageClass string) error {
	klog.V(4).Infof("STEP: Creating Bucket %s in region %s in COS %s storageClass %s", bucketName, region, cos, storageClass)
	s3Cli, err := client.NewS3Client(bxCli, cos, region)
	if err != nil {
		klog.Errorf("unable to create S3Client, err: %v", err)
		return err
	}

	_, err = s3Cli.S3Session.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(region + "-" + storageClass),
		},
	})
	if err != nil {
		klog.Errorf("unable to create bucket, err: %v", err)
		return err
	}

	err = s3Cli.S3Session.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		klog.Errorf("error while waiting for bucket, err: %v", err)
		return err
	}

	return nil
}

// Create Local object files
func createObjects() error {
	klog.V(4).Info("STEP: Create Required Files")
	var content string
	dir, err := os.MkdirTemp(".", "objects")
	if err != nil {
		klog.Errorf("unable to create temporary directory, err: %v", err)
		return err
	}

	ObjectsFolderName = dir
	for i := 0; i < numObjects; i++ {
		file, err := os.CreateTemp(ObjectsFolderName, "image-sync-*.txt")
		if err != nil {
			klog.Errorf("unable to create a temp file, err: %v", err)
			return err
		}
		defer file.Close()

		content = utils.GenerateRandomString(200)
		_, err = file.WriteString(content)
		if err != nil {
			klog.Errorf("error while writing in the file, err: %v", err)
			return err
		}
	}

	return nil
}

// Delete Temporarily created local object files and spec file
func deleteTempFiles() error {
	klog.V(4).Info("STEP: Delete created Files")
	specFolder := filepath.Dir(SpecFileName)
	klog.V(3).Infof("deleting spec folder:%s", specFolder)

	err := os.RemoveAll(specFolder)
	if err != nil {
		klog.Errorf("error while deleting spec folder, err: %v", err)
	}

	klog.V(3).Infof("deleting object folder:%s", ObjectsFolderName)
	err = os.RemoveAll(ObjectsFolderName)
	if err != nil {
		klog.Errorf("error while deleting object folder, err: %v", err)
	}

	return nil
}

// upload worker
func uploadWorker(s3Cli *client.S3Client, bucketName string, workerId int, filepaths <-chan string, results chan<- bool) {
	for filepath := range filepaths {
		fileName := strings.Split(filepath, "/")[len(strings.Split(filepath, "/"))-1]
		err := s3Cli.UploadObject(filepath, fileName, bucketName)
		if err != nil {
			klog.Errorf("file %s upload failed, err: %v", filepath, err)
			results <- false
		}
		results <- true
	}
}

// Upload object from local dir to s3 bucket
func uploadObjects(src pkg.Source) error {
	klog.V(4).Infof("STEP: Upload Objects to source Bucket %s", src.Bucket)
	var filePath string
	files, err := os.ReadDir(ObjectsFolderName)
	if err != nil {
		klog.Errorf("error while reading the directory, err: %v", err)
		return err
	}

	s3Cli, err := client.NewS3Client(bxCli, src.Cos, src.Region)
	if err != nil {
		klog.Errorf("unable to create S3Client, err: %v", err)
		return err
	}

	filepaths := make(chan string, len(files))
	results := make(chan bool, len(files))

	for w := 1; w <= numUploadWorkers; w++ {
		go uploadWorker(s3Cli, src.Bucket, w, filepaths, results)
	}

	for _, f := range files {
		filePath = ObjectsFolderName + "/" + f.Name()
		filepaths <- filePath
	}
	close(filepaths)

	for i := 1; i <= len(files); i++ {
		if !<-results {
			return errors.New("FAIL: Upload Objects failed")
		}
	}

	return nil
}

// Verify the copied Objects exists in the target bucket
func verifyBucketObjects(tgt pkg.TargetItem, cos string, files []fs.FileInfo, regex string) error {
	klog.V(4).Infof("STEP: Verify objects in Bucket %s", tgt.Bucket)

	s3Cli, err := client.NewS3Client(bxCli, cos, tgt.Region)
	if err != nil {
		klog.Errorf("unable to create S3Client, err: %v", err)
		return err
	}

	objects, err := s3Cli.SelectObjects(tgt.Bucket, regex)
	if err != nil {
		klog.Errorf("error while selecting objects, err: %v", err)
		return err
	}

	for _, f := range files {
		fileName := f.Name()
		res := false
		klog.V(3).Infof("Verifying object %s", fileName)

		for _, item := range objects {
			if item == fileName {
				res = true
				break
			}
		}
		if !res {
			klog.Errorf("object %s not found in the bucket %s", fileName, tgt.Bucket)
			return errors.New("ERROR: Object not found in the bucket ")
		}
	}

	return nil
}

// Verify objects copied from source bucket to dest buckets
func verifyObjectsCopied(spec []pkg.Spec) error {
	klog.V(4).Info("STEP: Verify Objects Copied to dest buckets")
	files, err := os.ReadDir(ObjectsFolderName)
	if err != nil {
		klog.Errorf("error while reading directory, err: %v", err)
		return err
	}
	fileInfos := make([]fs.FileInfo, 0, len(files))
	for _, entry := range files {
		fileInfo, err := entry.Info()
		if err != nil {
			return err
		}
		fileInfos = append(fileInfos, fileInfo)
	}

	for _, src := range spec {
		for _, tgt := range src.Target {
			err = verifyBucketObjects(tgt, src.Cos, fileInfos, src.Object)
			if err != nil {
				klog.Errorf("error while verifying bucket objects, err: %v", err)
				return err
			}
		}
	}

	return nil
}

// Create necessary resources to run the sync command
func createResources(spec []pkg.Spec) error {
	klog.V(4).Info("STEP: Create resources")
	err := createSpecFile(spec)
	if err != nil {
		return err
	}

	err = createObjects()
	if err != nil {
		return err
	}

	for _, src := range spec {
		err = createCOSInstance(src.Cos)
		if err != nil {
			return err
		}

		err = createBucket(src.Bucket, src.Cos, src.Region, src.StorageClass)
		if err != nil {
			return err
		}

		err = uploadObjects(src.Source)
		if err != nil {
			return err
		}

		for _, tgt := range src.Target {
			err = createBucket(tgt.Bucket, src.Cos, tgt.Region, tgt.StorageClass)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Delete the resources
func deleteResources(spec []pkg.Spec) error {
	klog.V(4).Info("STEP: Delete resources")
	for _, src := range spec {
		err := deleteCOSInstance(src.Cos)
		if err != nil {
			return err
		}
	}

	err := deleteTempFiles()
	if err != nil {
		return err
	}

	return nil
}

var _ = CMDDescribe("pvsadm image sync tests", func() {

	It("run with --help option", func() {
		status, stdout, stderr := runSyncCMD(
			"--help",
		)
		Expect(stderr).To(BeEmpty())
		Expect(status).To(BeZero())
		Expect(stdout).To(ContainSubstring("Examples:"))
	})

	framework.NegativeIt("run without spec-file flag", func() {
		status, _, stderr := runSyncCMD()
		Expect(status).NotTo(BeZero())
		Expect(stderr).To(ContainSubstring(`"spec-file" not set`))
	})

	framework.NegativeIt("run with yaml file that doesn't exist", func() {
		status, _, stderr := runSyncCMD("--spec-file", "fakefile.yaml")
		Expect(status).NotTo(BeZero())
		Expect(stderr).To(ContainSubstring(`no such file or directory`))
	})

	// Create a session to perform operations for e2e tests.
	BeforeEach(func() {
		bxCli, err = client.NewClientWithEnv(APIKey, client.DefaultEnv, debug)
		Expect(err).NotTo(HaveOccurred())
	})

	It("Copy Object Between Buckets", func() {

		specSlice := make([]pkg.Spec, 0)
		for i := 0; i < numSources; i++ {
			specSlice = append(specSlice, utils.GenerateSpec(numTargetsPerSource))
		}

		err = createResources(specSlice)
		Expect(err).NotTo(HaveOccurred())
		defer deleteResources(specSlice)

		status, _, _ := runSyncCMD("--spec-file", SpecFileName)
		Expect(status).To(BeZero())

		err = verifyObjectsCopied(specSlice)
		Expect(err).NotTo(HaveOccurred())
	})

})
